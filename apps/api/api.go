package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/perigee/perigee/apps/api/internal/orbit"
)

type SourceMetadata struct {
	SourceName     string  `json:"sourceName"`
	SourceURL      *string `json:"sourceUrl"`
	RetrievedAtUTC string  `json:"retrievedAtUtc"`
	DatasetType    string  `json:"datasetType"`
	SnapshotID     string  `json:"snapshotId"`
	License        *string `json:"license"`
	Caveat         string  `json:"caveat"`
}

type CatalogObject struct {
	CatalogID         string         `json:"catalogId"`
	Name              string         `json:"name"`
	CatalogSnapshotID string         `json:"catalogSnapshotId"`
	NORADID           *int           `json:"noradId"`
	SourceEpochUTC    string         `json:"sourceEpochUtc"`
	DatasetType       string         `json:"datasetType"`
	OrbitRegime       string         `json:"orbitRegime"`
	MeanAltitudeKM    float64        `json:"meanAltitudeKm"`
	DataQuality       string         `json:"dataQuality"`
	ElementModel      string         `json:"elementModel"`
	Source            SourceMetadata `json:"source"`
}

type CatalogFixture struct {
	CatalogSnapshotID string          `json:"catalogSnapshotId"`
	DatasetType       string          `json:"datasetType"`
	Snapshot          SourceMetadata  `json:"snapshot"`
	SourceEpochUTC    string          `json:"sourceEpochUtc"`
	Objects           []CatalogObject `json:"objects"`
}

type EncounterSummary struct {
	EncounterID        string  `json:"encounterId"`
	ObjectAID          string  `json:"objectAId"`
	ObjectBID          string  `json:"objectBId"`
	CatalogSnapshotID  string  `json:"catalogSnapshotId"`
	RunID              *string `json:"runId"`
	TCAUTC             string  `json:"tcaUtc"`
	MinimumSeparationM float64 `json:"minimumSeparationMeters"`
	RelativeSpeedMPS   float64 `json:"relativeSpeedMps"`
	SourceEpochUTC     string  `json:"sourceEpochUtc"`
	DatasetType        string  `json:"datasetType"`
	DataQuality        string  `json:"dataQuality"`
	Featured           bool    `json:"featured"`
}

type EncounterFixture struct {
	CatalogSnapshotID   string             `json:"catalogSnapshotId"`
	DatasetType         string             `json:"datasetType"`
	FeaturedEncounterID string             `json:"featuredEncounterId"`
	SourceEpochUTC      string             `json:"sourceEpochUtc"`
	Source              SourceMetadata     `json:"source"`
	Encounters          []EncounterSummary `json:"encounters"`
}

type CatalogPage struct {
	Items      []CatalogObject `json:"items"`
	NextCursor *string         `json:"nextCursor"`
}

type FeaturedEncounter struct {
	FeaturedEncounterID string           `json:"featuredEncounterId"`
	EncounterIDs        []string         `json:"encounterIds"`
	Snapshot            SourceMetadata   `json:"snapshot"`
	Encounter           EncounterSummary `json:"encounter"`
}

type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

type EncounterDetails struct {
	EncounterSummary
	ObjectA CatalogObject  `json:"objectA"`
	ObjectB CatalogObject  `json:"objectB"`
	Source  SourceMetadata `json:"source"`
	Caveats []string       `json:"caveats"`
}

type FrameMetadata struct {
	Frame        string `json:"frame"`
	PositionUnit string `json:"positionUnit"`
	VelocityUnit string `json:"velocityUnit"`
	DistanceUnit string `json:"distanceUnit"`
	FrameCaveat  string `json:"frameCaveat"`
}

type ObjectState struct {
	PositionXKM  float64 `json:"positionXKm"`
	PositionYKM  float64 `json:"positionYKm"`
	PositionZKM  float64 `json:"positionZKm"`
	VelocityXMPS float64 `json:"velocityXMps"`
	VelocityYMPS float64 `json:"velocityYMps"`
	VelocityZMPS float64 `json:"velocityZMps"`
}

type TrajectorySample struct {
	AtUTC            string      `json:"atUtc"`
	ObjectA          ObjectState `json:"objectA"`
	ObjectB          ObjectState `json:"objectB"`
	SeparationMeters float64     `json:"separationMeters"`
}

type Trajectory struct {
	EncounterID string             `json:"encounterId"`
	Frame       FrameMetadata      `json:"frame"`
	TCAUTC      string             `json:"tcaUtc"`
	StartUTC    string             `json:"startUtc"`
	EndUTC      string             `json:"endUtc"`
	Samples     []TrajectorySample `json:"samples"`
	Source      SourceMetadata     `json:"source"`
}

type ManeuverRequest struct {
	ManeuverTimeUTC     string   `json:"maneuverTimeUtc"`
	AlongTrackDeltaVMPS *float64 `json:"alongTrackDeltaVMps"`
}

type OMMImportSource struct {
	Name           string `json:"name"`
	URL            string `json:"url,omitempty"`
	SnapshotID     string `json:"snapshotId"`
	RetrievedAtUTC string `json:"retrievedAtUtc"`
}

type OMMImportRequest struct {
	Source             OMMImportSource `json:"source"`
	MaxEpochAgeSeconds int64           `json:"maxEpochAgeSeconds"`
	Records            json.RawMessage `json:"records"`
}

type OMMImportResponse struct {
	SnapshotID string                   `json:"snapshotId"`
	Source     orbit.CatalogSource      `json:"source"`
	Records    []orbit.OMMCatalogRecord `json:"records"`
	Caveats    []string                 `json:"caveats"`
}

type ManeuverMetrics struct {
	TCAUTC                  string  `json:"tcaUtc"`
	MinimumSeparationMeters float64 `json:"minimumSeparationMeters"`
	RelativeSpeedMPS        float64 `json:"relativeSpeedMps"`
}

type ManeuverResult struct {
	ScenarioID          string          `json:"scenarioId"`
	EncounterID         string          `json:"encounterId"`
	ManeuverTimeUTC     string          `json:"maneuverTimeUtc"`
	AlongTrackDeltaVMPS float64         `json:"alongTrackDeltaVMps"`
	OriginalMetrics     ManeuverMetrics `json:"originalMetrics"`
	SimulatedMetrics    ManeuverMetrics `json:"simulatedMetrics"`
	MissDistanceDeltaM  float64         `json:"missDistanceDeltaMeters"`
	OriginalTrajectory  Trajectory      `json:"originalTrajectory"`
	SimulatedTrajectory Trajectory      `json:"simulatedTrajectory"`
	Limitations         []string        `json:"limitations"`
	DatasetType         string          `json:"datasetType"`
}

type API struct {
	catalog    CatalogFixture
	encounters EncounterFixture
	objects    map[string]CatalogObject
	byID       map[string]EncounterSummary

	mu                   sync.Mutex
	ommSnapshots         map[string]ommSnapshot
	screenings           map[string]*screeningRun
	screeningQueue       chan screeningTask
	screeningWorkersOnce sync.Once
}

func NewHandler() (http.Handler, error) {
	dir, err := findFixtureDir()
	if err != nil {
		return nil, err
	}
	return NewHandlerWithFixtureDir(dir)
}

func NewHandlerWithFixtureDir(dir string) (http.Handler, error) {
	var catalog CatalogFixture
	if err := readJSON(filepath.Join(dir, "catalog.json"), &catalog); err != nil {
		return nil, fmt.Errorf("read catalog fixture: %w", err)
	}
	var encounters EncounterFixture
	if err := readJSON(filepath.Join(dir, "encounters.json"), &encounters); err != nil {
		return nil, fmt.Errorf("read encounters fixture: %w", err)
	}

	api := &API{
		catalog:        catalog,
		encounters:     encounters,
		objects:        make(map[string]CatalogObject, len(catalog.Objects)),
		byID:           make(map[string]EncounterSummary, len(encounters.Encounters)),
		ommSnapshots:   make(map[string]ommSnapshot),
		screenings:     make(map[string]*screeningRun),
		screeningQueue: make(chan screeningTask, screeningQueueSize),
	}
	for _, object := range catalog.Objects {
		api.objects[object.CatalogID] = object
	}
	for _, encounter := range encounters.Encounters {
		api.byID[encounter.EncounterID] = encounter
	}
	return api, nil
}

func findFixtureDir() (string, error) {
	if dir := os.Getenv("PERIGEE_FIXTURES_DIR"); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "catalog.json")); err == nil {
			return dir, nil
		}
		return "", fmt.Errorf("PERIGEE_FIXTURES_DIR %q does not contain catalog.json", dir)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "data", "fixtures")
		if _, err := os.Stat(filepath.Join(candidate, "catalog.json")); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("could not locate data/fixtures/catalog.json; set PERIGEE_FIXTURES_DIR")
}

func readJSON(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(target); err != nil {
		return err
	}
	return nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/screenings/") {
		a.serveScreeningRun(w, r)
		return
	}
	postAllowed := r.Method == http.MethodPost &&
		(strings.HasPrefix(r.URL.Path, "/api/encounters/") || r.URL.Path == "/api/catalog/omm" || r.URL.Path == "/api/screenings")
	if r.Method != http.MethodGet && !postAllowed {
		allow := "GET"
		if r.URL.Path == "/api/catalog/omm" {
			allow = "POST"
		} else if r.URL.Path == "/api/screenings" {
			allow = "POST"
		} else if strings.HasPrefix(r.URL.Path, "/api/encounters/") {
			allow = "GET, POST"
		}
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported", nil)
		return
	}

	switch {
	case r.URL.Path == "/api/health":
		if r.Method != http.MethodGet {
			break
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok", "service": "perigee-api", "version": "0.1.0",
			"timestampUtc": time.Now().UTC().Format(time.RFC3339), "fixtureMode": true,
		})
		return
	case r.URL.Path == "/api/catalog":
		if r.Method == http.MethodGet {
			a.listCatalog(w, r)
			return
		}
	case r.URL.Path == "/api/catalog/omm":
		if r.Method == http.MethodPost {
			a.importOMMCatalog(w, r)
			return
		}
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported", nil)
		return
	case r.URL.Path == "/api/screenings":
		if r.Method == http.MethodPost {
			a.createScreening(w, r)
			return
		}
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported", nil)
		return
	case r.URL.Path == "/api/demo/featured":
		if r.Method == http.MethodGet {
			a.featured(w)
			return
		}
	case strings.HasPrefix(r.URL.Path, "/api/catalog/"):
		if r.Method == http.MethodGet {
			a.getCatalogObject(w, strings.TrimPrefix(r.URL.Path, "/api/catalog/"))
			return
		}
	case strings.HasPrefix(r.URL.Path, "/api/encounters/"):
		a.handleEncounter(w, r)
		return
	default:
		writeError(w, http.StatusNotFound, "not_found", "route not found", nil)
		return
	}
	w.Header().Set("Allow", "GET")
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported", nil)
}

func (a *API) importOMMCatalog(w http.ResponseWriter, r *http.Request) {
	const maxRequestBytes = 4 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	defer r.Body.Close()

	var request OMMImportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "OMM import request exceeds the 4 MiB limit", nil)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_omm_request", "request body must be valid OMM import JSON", map[string]any{"reason": err.Error()})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "OMM import request exceeds the 4 MiB limit", nil)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_omm_request", "request body must contain one JSON value", nil)
		return
	}

	retrievedAt, err := parseUTC(request.Source.RetrievedAtUTC)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_source_metadata", "source.retrievedAtUtc must be RFC3339 UTC ending in Z", nil)
		return
	}
	maxDurationSeconds := int64(time.Duration(1<<63-1) / time.Second)
	if request.MaxEpochAgeSeconds <= 0 || request.MaxEpochAgeSeconds > maxDurationSeconds {
		writeError(w, http.StatusBadRequest, "invalid_epoch_age", "maxEpochAgeSeconds must be positive and fit in a duration", nil)
		return
	}

	source := orbit.CatalogSource{
		Name:           request.Source.Name,
		URL:            request.Source.URL,
		SnapshotID:     request.Source.SnapshotID,
		RetrievedAtUTC: retrievedAt,
	}
	records, err := orbit.ReadOMMCatalog(
		bytes.NewReader(request.Records),
		source,
		time.Duration(request.MaxEpochAgeSeconds)*time.Second,
	)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_omm_catalog", "OMM catalog failed validation", map[string]any{"reason": err.Error()})
		return
	}
	if err := a.registerOMMSnapshot(source, records); err != nil {
		writeError(w, http.StatusConflict, "snapshot_id_conflict", "snapshot ID is already registered with different contents", map[string]any{"reason": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, OMMImportResponse{
		SnapshotID: source.SnapshotID,
		Source:     source,
		Records:    records,
		Caveats: []string{
			"Validated records are retained in a bounded in-memory snapshot cache for screening; they are cleared on API restart and are not merged into /api/catalog or fixture encounter routes.",
			"Propagation support indicates compatibility with the selected near-Earth SGP4 adapter, not accuracy or collision risk.",
			"Mean-motion derivative interpretation follows the selected SGP4 dependency and must be checked against the source convention.",
		},
	})
}

func (a *API) listCatalog(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be an integer from 1 to 100", nil)
			return
		}
		limit = value
	}
	offset := 0
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			writeError(w, http.StatusBadRequest, "invalid_cursor", "cursor must be a non-negative integer", nil)
			return
		}
		offset = value
	}
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("query")))
	items := append([]CatalogObject(nil), a.catalog.Objects...)
	sort.Slice(items, func(i, j int) bool { return items[i].CatalogID < items[j].CatalogID })
	if query != "" {
		filtered := items[:0]
		for _, object := range items {
			if strings.Contains(strings.ToLower(object.CatalogID+" "+object.Name+" "+object.OrbitRegime), query) {
				filtered = append(filtered, object)
			}
		}
		items = filtered
	}
	if offset > len(items) {
		offset = len(items)
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[offset:end]
	next := (*string)(nil)
	if end < len(items) {
		cursor := strconv.Itoa(end)
		next = &cursor
	}
	writeJSON(w, http.StatusOK, CatalogPage{Items: pageItems, NextCursor: next})
}

func (a *API) getCatalogObject(w http.ResponseWriter, id string) {
	if object, ok := a.objects[id]; ok {
		writeJSON(w, http.StatusOK, object)
		return
	}
	writeError(w, http.StatusNotFound, "catalog_object_not_found", "catalog object was not found", map[string]any{"catalogId": id})
}

func (a *API) featured(w http.ResponseWriter) {
	var featured *EncounterSummary
	ids := make([]string, 0, len(a.encounters.Encounters))
	for i := range a.encounters.Encounters {
		encounter := a.encounters.Encounters[i]
		ids = append(ids, encounter.EncounterID)
		if encounter.EncounterID == a.encounters.FeaturedEncounterID {
			copy := encounter
			featured = &copy
		}
	}
	if featured == nil {
		writeError(w, http.StatusInternalServerError, "fixture_invalid", "featured encounter is missing from fixture", nil)
		return
	}
	writeJSON(w, http.StatusOK, FeaturedEncounter{
		FeaturedEncounterID: featured.EncounterID, EncounterIDs: ids,
		Snapshot: a.encounters.Source, Encounter: *featured,
	})
}

func (a *API) handleEncounter(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/encounters/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "route not found", nil)
		return
	}
	encounter, ok := a.byID[parts[0]]
	if !ok {
		writeError(w, http.StatusNotFound, "encounter_not_found", "encounter was not found", map[string]any{"encounterId": parts[0]})
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		objectA, okA := a.objects[encounter.ObjectAID]
		objectB, okB := a.objects[encounter.ObjectBID]
		if !okA || !okB {
			writeError(w, http.StatusInternalServerError, "fixture_invalid", "encounter references an unknown catalog object", nil)
			return
		}
		writeJSON(w, http.StatusOK, EncounterDetails{
			EncounterSummary: encounter, ObjectA: objectA, ObjectB: objectB,
			Source: a.encounters.Source,
			Caveats: []string{
				"Synthetic educational fixture; not derived from real conjunction data.",
				"Minimum separation is a miss distance, not a collision probability or operational prediction.",
			},
		})
		return
	}
	if len(parts) == 2 && parts[1] == "trajectory" && r.Method == http.MethodGet {
		trajectory, status, err := a.trajectory(encounter, r.URL.Query(), 0, "")
		if err != nil {
			writeError(w, status, "invalid_trajectory_request", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, trajectory)
		return
	}
	if len(parts) == 2 && parts[1] == "maneuvers" && r.Method == http.MethodPost {
		a.maneuver(w, r, encounter)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported", nil)
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found", nil)
}

func (a *API) trajectory(encounter EncounterSummary, query url.Values, maneuverMPS float64, maneuverTime string) (Trajectory, int, error) {
	tca, err := parseUTC(encounter.TCAUTC)
	if err != nil {
		return Trajectory{}, http.StatusInternalServerError, errors.New("fixture contains an invalid TCA timestamp")
	}
	start := tca.Add(-15 * time.Minute)
	end := tca.Add(15 * time.Minute)
	if raw := query.Get("startUtc"); raw != "" {
		start, err = parseUTC(raw)
		if err != nil {
			return Trajectory{}, http.StatusUnprocessableEntity, errors.New("startUtc must be an RFC3339 UTC timestamp ending in Z")
		}
	}
	if raw := query.Get("endUtc"); raw != "" {
		end, err = parseUTC(raw)
		if err != nil {
			return Trajectory{}, http.StatusUnprocessableEntity, errors.New("endUtc must be an RFC3339 UTC timestamp ending in Z")
		}
	}
	if !start.Before(end) {
		return Trajectory{}, http.StatusUnprocessableEntity, errors.New("startUtc must be before endUtc")
	}
	count := 61
	if raw := query.Get("samples"); raw != "" {
		count, err = strconv.Atoi(raw)
		if err != nil || count < 2 || count > 241 {
			return Trajectory{}, http.StatusUnprocessableEntity, errors.New("samples must be an integer from 2 to 241")
		}
	}
	maneuverAt := time.Time{}
	if maneuverTime != "" {
		maneuverAt, err = parseUTC(maneuverTime)
		if err != nil {
			return Trajectory{}, http.StatusBadRequest, errors.New("maneuverTimeUtc must be an RFC3339 UTC timestamp ending in Z")
		}
	}

	step := end.Sub(start) / time.Duration(count-1)
	samples := make([]TrajectorySample, 0, count)
	for i := 0; i < count; i++ {
		at := start.Add(time.Duration(i) * step)
		if i == count-1 {
			at = end
		}
		secondsFromTCA := at.Sub(tca).Seconds()
		maneuverElapsed := 0.0
		maneuverVelocity := 0.0
		if !maneuverAt.IsZero() && at.After(maneuverAt) {
			maneuverElapsed = at.Sub(maneuverAt).Seconds()
		}
		if !maneuverAt.IsZero() && !at.Before(maneuverAt) {
			maneuverVelocity = maneuverMPS
		}
		objectA := ObjectState{
			PositionXKM:  6878 + maneuverMPS*maneuverElapsed/1000,
			PositionYKM:  7.6 * secondsFromTCA,
			PositionZKM:  0,
			VelocityXMPS: maneuverVelocity,
			VelocityYMPS: 7600,
			VelocityZMPS: 0,
		}
		objectB := ObjectState{
			PositionXKM:  6878 + encounter.MinimumSeparationM/1000,
			PositionYKM:  7.6 * secondsFromTCA,
			PositionZKM:  encounter.RelativeSpeedMPS * secondsFromTCA / 1000,
			VelocityXMPS: 0,
			VelocityYMPS: 7600,
			VelocityZMPS: encounter.RelativeSpeedMPS,
		}
		dx := (objectB.PositionXKM - objectA.PositionXKM) * 1000
		dz := (objectB.PositionZKM - objectA.PositionZKM) * 1000
		distance := hypot(dx, dz)
		samples = append(samples, TrajectorySample{
			AtUTC: at.UTC().Format(time.RFC3339), ObjectA: objectA, ObjectB: objectB,
			SeparationMeters: distance,
		})
	}
	return Trajectory{
		EncounterID: encounter.EncounterID,
		Frame: FrameMetadata{
			Frame: "ECI", PositionUnit: "km", VelocityUnit: "m/s", DistanceUnit: "m",
			FrameCaveat: "Synthetic local linear encounter frame labeled ECI for the demo contract; not a physically propagated inertial ephemeris.",
		},
		TCAUTC: encounter.TCAUTC, StartUTC: start.UTC().Format(time.RFC3339),
		EndUTC: end.UTC().Format(time.RFC3339), Samples: samples, Source: a.encounters.Source,
	}, http.StatusOK, nil
}

func (a *API) maneuver(w http.ResponseWriter, r *http.Request, encounter EncounterSummary) {
	var request ManeuverRequest
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON with only documented fields", map[string]any{"reason": err.Error()})
		return
	}
	maneuverAt, err := parseUTC(request.ManeuverTimeUTC)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_maneuver_time", "maneuverTimeUtc must be an RFC3339 UTC timestamp ending in Z", nil)
		return
	}
	if request.AlongTrackDeltaVMPS == nil {
		writeError(w, http.StatusBadRequest, "invalid_delta_v", "alongTrackDeltaVMps is required", nil)
		return
	}
	deltaV := *request.AlongTrackDeltaVMPS
	if deltaV < -0.5 || deltaV > 0.5 {
		writeError(w, http.StatusBadRequest, "invalid_delta_v", "alongTrackDeltaVMps must be between -0.5 and 0.5 m/s", nil)
		return
	}
	tca, err := parseUTC(encounter.TCAUTC)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "fixture_invalid", "encounter TCA is invalid", nil)
		return
	}
	if !maneuverAt.Before(tca) || maneuverAt.Before(tca.Add(-15*time.Minute)) {
		writeError(w, http.StatusBadRequest, "invalid_maneuver_time", "maneuverTimeUtc must be within the 15-minute synthetic trajectory window before TCA", nil)
		return
	}
	original, status, err := a.trajectory(encounter, url.Values{}, 0, "")
	if err != nil {
		writeError(w, status, "invalid_trajectory_request", err.Error(), nil)
		return
	}
	simulated, status, err := a.trajectory(encounter, url.Values{}, deltaV, request.ManeuverTimeUTC)
	if err != nil {
		writeError(w, status, "invalid_trajectory_request", err.Error(), nil)
		return
	}
	simulatedMetrics := closestMetrics(simulated.Samples, encounter.RelativeSpeedMPS, deltaV)
	result := ManeuverResult{
		ScenarioID:          scenarioID(encounter.EncounterID, request.ManeuverTimeUTC, deltaV),
		EncounterID:         encounter.EncounterID,
		ManeuverTimeUTC:     request.ManeuverTimeUTC,
		AlongTrackDeltaVMPS: deltaV,
		OriginalMetrics: ManeuverMetrics{
			TCAUTC:                  encounter.TCAUTC,
			MinimumSeparationMeters: encounter.MinimumSeparationM,
			RelativeSpeedMPS:        encounter.RelativeSpeedMPS,
		},
		SimulatedMetrics:    simulatedMetrics,
		MissDistanceDeltaM:  simulatedMetrics.MinimumSeparationMeters - encounter.MinimumSeparationM,
		OriginalTrajectory:  original,
		SimulatedTrajectory: simulated,
		Limitations: []string{
			"This deterministic synthetic demo applies a constant linear position offset after an instantaneous velocity change; it is not orbital dynamics.",
			"The maneuver is an educational what-if only, not operational guidance or a guarantee of collision avoidance.",
		},
		DatasetType: encounter.DatasetType,
	}
	writeJSON(w, http.StatusOK, result)
}

func parseUTC(value string) (time.Time, error) {
	if !strings.HasSuffix(value, "Z") {
		return time.Time{}, errors.New("timestamp must end in Z")
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	_, offset := parsed.Zone()
	if offset != 0 {
		return time.Time{}, errors.New("timestamp must be UTC")
	}
	return parsed.UTC(), nil
}

func hypot(x, y float64) float64 {
	return math.Hypot(x, y)
}

func closestMetrics(samples []TrajectorySample, relativeSpeed, deltaV float64) ManeuverMetrics {
	best := samples[0]
	for _, sample := range samples[1:] {
		if sample.SeparationMeters < best.SeparationMeters {
			best = sample
		}
	}
	return ManeuverMetrics{
		TCAUTC: best.AtUTC, MinimumSeparationMeters: best.SeparationMeters,
		RelativeSpeedMPS: hypot(relativeSpeed, deltaV),
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	writeJSON(w, status, ErrorResponse{Code: code, Message: message, Details: details})
}

func decodeBody(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 64*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return err
		}
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func scenarioID(encounterID, maneuverTime string, deltaV float64) string {
	raw := fmt.Sprintf("%s|%s|%.6f", encounterID, maneuverTime, deltaV)
	sum := sha256.Sum256([]byte(raw))
	return "scenario-" + hex.EncodeToString(sum[:6])
}
