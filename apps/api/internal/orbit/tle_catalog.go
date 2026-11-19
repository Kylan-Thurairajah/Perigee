package orbit

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/akhenakh/sgp4"
)

type CatalogRecord struct {
	SatelliteNumber       int
	Name                  string
	EpochUTC              time.Time
	Source                CatalogSource
	DataQuality           string
	TLELine1              string
	TLELine2              string
	PropagationSupported  bool
	PropagationLimitation string
}

func ReadTLECatalog(reader io.Reader, source CatalogSource, maxEpochAge time.Duration) ([]CatalogRecord, error) {
	if reader == nil {
		return nil, fmt.Errorf("TLE catalog reader is required")
	}
	var err error
	source, err = validateCatalogSource(source, maxEpochAge, "TLE")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	var records []CatalogRecord
	seen := make(map[int]int)
	var name, line1 string
	state := "start"
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch state {
		case "start":
			if isTLELine(line, '1') {
				line1 = line
				state = "line2"
			} else if isTLELine(line, '2') {
				return nil, fmt.Errorf("line %d: TLE line 2 appeared before line 1", lineNumber)
			} else {
				name = line
				state = "line1"
			}
		case "line1":
			if !isTLELine(line, '1') {
				return nil, fmt.Errorf("line %d: expected TLE line 1 after object name", lineNumber)
			}
			line1 = line
			state = "line2"
		case "line2":
			if !isTLELine(line, '2') {
				return nil, fmt.Errorf("line %d: expected TLE line 2", lineNumber)
			}
			record, err := parseCatalogRecord(name, line1, line, source, maxEpochAge)
			if err != nil {
				return nil, fmt.Errorf("lines %d-%d: %w", lineNumber-1, lineNumber, err)
			}
			if previousLine, exists := seen[record.SatelliteNumber]; exists {
				return nil, fmt.Errorf("line %d: duplicate catalog number %d (first seen at line %d)", lineNumber, record.SatelliteNumber, previousLine)
			}
			seen[record.SatelliteNumber] = lineNumber - 1
			records = append(records, record)
			name, line1, state = "", "", "start"
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read TLE catalog: %w", err)
	}
	if state != "start" {
		return nil, fmt.Errorf("incomplete TLE record at end of catalog")
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("TLE catalog contains no records")
	}
	return records, nil
}

func isTLELine(line string, lineNumber byte) bool {
	return len(line) >= 2 && line[0] == lineNumber && line[1] == ' '
}

func parseCatalogRecord(name, line1, line2 string, source CatalogSource, maxEpochAge time.Duration) (CatalogRecord, error) {
	tle, err := sgp4.ParseTLELines([]string{line1, line2})
	if err != nil {
		return CatalogRecord{}, fmt.Errorf("parse TLE: %w", err)
	}
	epoch := tle.EpochTime().UTC()
	if epoch.After(source.RetrievedAtUTC) {
		return CatalogRecord{}, fmt.Errorf("TLE epoch %s is after source retrieval time %s", epoch.Format(time.RFC3339), source.RetrievedAtUTC.Format(time.RFC3339))
	}
	if name != "" {
		tle.Name = name
	}
	record := CatalogRecord{
		SatelliteNumber: tle.SatelliteNumber,
		Name:            tle.Name,
		EpochUTC:        epoch,
		Source:          source,
		DataQuality:     "public_snapshot",
		TLELine1:        line1,
		TLELine2:        line2,
	}
	if source.RetrievedAtUTC.Sub(epoch) > maxEpochAge {
		record.DataQuality = "stale"
	}
	if _, err := tle.Initialize(); err != nil {
		record.PropagationLimitation = err.Error()
		return record, nil
	}
	record.PropagationSupported = true
	return record, nil
}
