package orbit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/akhenakh/sgp4"
)

const (
	maxOMMCatalogBytes   = 16 << 20
	maxOMMCatalogRecords = 10000
)

type OMMCatalogRecord struct {
	Element               sgp4.OMM        `json:"element"`
	RawElement            json.RawMessage `json:"rawElement"`
	SatelliteNumber       int             `json:"noradCatId"`
	EpochUTC              time.Time       `json:"epochUtc"`
	Source                CatalogSource   `json:"source"`
	DataQuality           string          `json:"dataQuality"`
	MissingFields         []string        `json:"missingFields,omitempty"`
	PropagationSupported  bool            `json:"propagationSupported"`
	PropagationLimitation string          `json:"propagationLimitation,omitempty"`
}

func ReadOMMCatalog(reader io.Reader, source CatalogSource, maxEpochAge time.Duration) ([]OMMCatalogRecord, error) {
	if reader == nil {
		return nil, fmt.Errorf("OMM catalog reader is required")
	}
	var err error
	source, err = validateCatalogSource(source, maxEpochAge, "OMM")
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(reader, maxOMMCatalogBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read OMM catalog: %w", err)
	}
	if len(data) > maxOMMCatalogBytes {
		return nil, fmt.Errorf("OMM catalog exceeds the %d-byte limit", maxOMMCatalogBytes)
	}
	data = bytes.TrimSpace(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}))
	if len(data) == 0 {
		return nil, fmt.Errorf("OMM catalog is empty")
	}

	var rawRecords []json.RawMessage
	switch data[0] {
	case '[':
		if err := json.Unmarshal(data, &rawRecords); err != nil {
			return nil, fmt.Errorf("decode OMM catalog JSON: %w", err)
		}
	case '{':
		rawRecords = []json.RawMessage{data}
	default:
		return nil, fmt.Errorf("OMM catalog must be a JSON object or array")
	}
	if len(rawRecords) == 0 {
		return nil, fmt.Errorf("OMM catalog contains no records")
	}
	if len(rawRecords) > maxOMMCatalogRecords {
		return nil, fmt.Errorf("OMM catalog exceeds the %d-record limit", maxOMMCatalogRecords)
	}

	records := make([]OMMCatalogRecord, 0, len(rawRecords))
	seen := make(map[int]int, len(rawRecords))
	for index, raw := range rawRecords {
		record, err := parseOMMCatalogRecord(raw, source, maxEpochAge)
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", index+1, err)
		}
		if first, exists := seen[record.SatelliteNumber]; exists {
			return nil, fmt.Errorf("record %d: duplicate catalog number %d (first seen at record %d)", index+1, record.SatelliteNumber, first)
		}
		seen[record.SatelliteNumber] = index + 1
		records = append(records, record)
	}
	return records, nil
}

func parseOMMCatalogRecord(raw json.RawMessage, source CatalogSource, maxEpochAge time.Duration) (OMMCatalogRecord, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		if err != nil {
			return OMMCatalogRecord{}, fmt.Errorf("decode OMM record: %w", err)
		}
		return OMMCatalogRecord{}, fmt.Errorf("OMM record must be a JSON object")
	}
	for _, field := range []string{
		"OBJECT_NAME", "EPOCH", "MEAN_MOTION", "ECCENTRICITY", "INCLINATION",
		"RA_OF_ASC_NODE", "ARG_OF_PERICENTER", "MEAN_ANOMALY", "NORAD_CAT_ID",
	} {
		value, exists := fields[field]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return OMMCatalogRecord{}, fmt.Errorf("required OMM field %s is missing", field)
		}
	}

	var element sgp4.OMM
	if err := json.Unmarshal(raw, &element); err != nil {
		return OMMCatalogRecord{}, fmt.Errorf("decode OMM fields: %w", err)
	}
	if strings.TrimSpace(element.ObjectName) == "" {
		return OMMCatalogRecord{}, fmt.Errorf("OBJECT_NAME must not be empty")
	}
	if element.NoradCatID <= 0 {
		return OMMCatalogRecord{}, fmt.Errorf("NORAD_CAT_ID must be positive")
	}
	if err := validateOMMElements(element); err != nil {
		return OMMCatalogRecord{}, err
	}

	epoch, err := parseOMMEpoch(element.EpochStr, element.TimeSystem)
	if err != nil {
		return OMMCatalogRecord{}, fmt.Errorf("parse EPOCH: %w", err)
	}
	if epoch.After(source.RetrievedAtUTC) {
		return OMMCatalogRecord{}, fmt.Errorf("OMM epoch %s is after source retrieval time %s", epoch.Format(time.RFC3339Nano), source.RetrievedAtUTC.Format(time.RFC3339Nano))
	}

	missing := missingOMMFields(fields)
	record := OMMCatalogRecord{
		Element:         element,
		RawElement:      append(json.RawMessage(nil), raw...),
		SatelliteNumber: element.NoradCatID,
		EpochUTC:        epoch,
		Source:          source,
		DataQuality:     "public_snapshot",
		MissingFields:   missing,
	}
	if source.RetrievedAtUTC.Sub(epoch) > maxEpochAge {
		record.DataQuality = "stale"
	}

	propagationMissing := missingOMMPropagationFields(fields)
	if len(propagationMissing) > 0 {
		record.PropagationLimitation = "OMM is missing fields required for propagation: " + strings.Join(propagationMissing, ", ")
		return record, nil
	}
	_, err = NewPropagatorFromOMM(element)
	if err != nil {
		record.PropagationLimitation = err.Error()
		return record, nil
	}
	record.PropagationSupported = true
	return record, nil
}

func validateOMMElements(element sgp4.OMM) error {
	values := []float64{
		element.MeanMotion, element.Eccentricity, element.Inclination,
		element.RAOfAscNode, element.ArgOfPericenter, element.MeanAnomaly,
		element.BStar, element.MeanMotionDot, element.MeanMotionDDot,
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("OMM orbital elements must be finite numbers")
		}
	}
	if element.MeanMotion <= 0 || element.MeanMotion >= 100 {
		return fmt.Errorf("MEAN_MOTION must be greater than 0 and less than 100 rev/day")
	}
	if element.Eccentricity < 0 || element.Eccentricity >= 1 {
		return fmt.Errorf("ECCENTRICITY must be in [0, 1)")
	}
	if element.Inclination < 0 || element.Inclination > 180 {
		return fmt.Errorf("INCLINATION must be in [0, 180] degrees")
	}
	if element.EphemerisType < 0 || element.EphemerisType > 9 {
		return fmt.Errorf("EPHEMERIS_TYPE must be in [0, 9]")
	}
	if element.ElementSetNo < 0 {
		return fmt.Errorf("ELEMENT_SET_NO must not be negative")
	}
	if element.RevAtEpoch < 0 {
		return fmt.Errorf("REV_AT_EPOCH must not be negative")
	}
	if len(element.ClassificationType) > 1 {
		return fmt.Errorf("CLASSIFICATION_TYPE must be a single character")
	}
	for _, field := range []struct {
		name  string
		value float64
	}{
		{name: "RA_OF_ASC_NODE", value: element.RAOfAscNode},
		{name: "ARG_OF_PERICENTER", value: element.ArgOfPericenter},
		{name: "MEAN_ANOMALY", value: element.MeanAnomaly},
	} {
		if field.value < 0 || field.value > 360 {
			return fmt.Errorf("%s must be in [0, 360] degrees", field.name)
		}
	}
	return nil
}

func missingOMMFields(fields map[string]json.RawMessage) []string {
	optionalFields := []string{
		"OBJECT_ID", "CENTER_NAME", "REF_FRAME", "TIME_SYSTEM", "MEAN_ELEMENT_THEORY",
		"EPHEMERIS_TYPE", "CLASSIFICATION_TYPE", "ELEMENT_SET_NO", "REV_AT_EPOCH",
		"BSTAR", "MEAN_MOTION_DOT", "MEAN_MOTION_DDOT",
	}
	return missingFields(fields, optionalFields)
}

func missingOMMPropagationFields(fields map[string]json.RawMessage) []string {
	requiredForPropagation := []string{
		"CENTER_NAME", "REF_FRAME", "TIME_SYSTEM", "MEAN_ELEMENT_THEORY",
		"EPHEMERIS_TYPE", "BSTAR", "MEAN_MOTION_DOT", "MEAN_MOTION_DDOT",
	}
	return missingFields(fields, requiredForPropagation)
}

func missingFields(fields map[string]json.RawMessage, names []string) []string {
	var missing []string
	for _, field := range names {
		value, exists := fields[field]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			missing = append(missing, field)
		}
	}
	return missing
}

func parseOMMEpoch(value, timeSystem string) (time.Time, error) {
	value = strings.TrimSpace(value)
	epoch, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return epoch.UTC(), nil
	}
	if !strings.EqualFold(strings.TrimSpace(timeSystem), "UTC") {
		return time.Time{}, fmt.Errorf("epoch without a timezone requires TIME_SYSTEM UTC")
	}
	for _, layout := range []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
	} {
		if epoch, parseErr := time.ParseInLocation(layout, value, time.UTC); parseErr == nil {
			return epoch.UTC(), nil
		}
	}
	return time.Time{}, err
}

func ommPropagationLimitation(element sgp4.OMM) string {
	for _, requirement := range []struct {
		field string
		value string
		want  string
	}{
		{"CENTER_NAME", element.CenterName, "EARTH"},
		{"REF_FRAME", element.RefFrame, "TEME"},
		{"TIME_SYSTEM", element.TimeSystem, "UTC"},
		{"MEAN_ELEMENT_THEORY", element.MeanElementTheory, "SGP4"},
	} {
		if !strings.EqualFold(strings.TrimSpace(requirement.value), requirement.want) {
			return fmt.Sprintf("%s must be %s (got %q)", requirement.field, requirement.want, requirement.value)
		}
	}
	if element.EphemerisType != 0 && element.EphemerisType != 2 {
		return fmt.Sprintf("EPHEMERIS_TYPE %d is not supported by the selected SGP4 adapter", element.EphemerisType)
	}
	return ""
}
