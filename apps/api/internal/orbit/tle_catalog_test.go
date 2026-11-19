package orbit

import (
	"strings"
	"testing"
	"time"

	"github.com/akhenakh/sgp4"
)

const (
	norad5Line1 = "1 00005U 58002B   00179.78495062  .00000023  00000-0  28098-4 0  4753"
	norad5Line2 = "2 00005  34.2682 348.7242 1859667 331.7664  19.3264 10.82419157413667"
)

func TestReadTLECatalogParsesNamedAndUnnamedRecords(t *testing.T) {
	retrieved := time.Date(2000, time.July, 1, 0, 0, 0, 0, time.FixedZone("local", -5*60*60))
	content := "VANGUARD 1\n" + norad5Line1 + "\n" + norad5Line2 + "\n"
	records, err := ReadTLECatalog(strings.NewReader(content), testCatalogSource(retrieved), 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadTLECatalog: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("record count = %d, want 1", len(records))
	}
	record := records[0]
	if record.SatelliteNumber != 5 || record.Name != "VANGUARD 1" {
		t.Fatalf("unexpected identity: %+v", record)
	}
	if !record.PropagationSupported || record.DataQuality != "public_snapshot" {
		t.Fatalf("unexpected validation status: %+v", record)
	}
	if record.Source.RetrievedAtUTC.Location() != time.UTC {
		t.Fatalf("retrieval time location = %v, want UTC", record.Source.RetrievedAtUTC.Location())
	}

	records, err = ReadTLECatalog(strings.NewReader(norad5Line1+"\n"+norad5Line2), testCatalogSource(retrieved), 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadTLECatalog without name: %v", err)
	}
	if records[0].Name != "" {
		t.Fatalf("unnamed TLE name = %q, want empty", records[0].Name)
	}
}

func TestReadTLECatalogMarksOldEpochStale(t *testing.T) {
	retrieved := time.Date(2000, time.July, 10, 0, 0, 0, 0, time.UTC)
	records, err := ReadTLECatalog(strings.NewReader(norad5Line1+"\n"+norad5Line2), testCatalogSource(retrieved), 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadTLECatalog: %v", err)
	}
	if records[0].DataQuality != "stale" {
		t.Fatalf("DataQuality = %q, want stale", records[0].DataQuality)
	}
}

func TestReadTLECatalogRetainsUnsupportedOrbitWithLimitation(t *testing.T) {
	tle := &sgp4.TLE{
		Name:            "GEO TEST",
		SatelliteNumber: 99999,
		Classification:  'U',
		EpochYear:       2000,
		EpochDay:        179.5,
		EphemerisType:   0,
		Inclination:     0,
		RightAscension:  180,
		Eccentricity:    0.0001,
		ArgOfPerigee:    0,
		MeanAnomaly:     0,
		MeanMotion:      1.0027,
	}
	lines, err := tle.TLELines()
	if err != nil {
		t.Fatalf("TLELines: %v", err)
	}
	content := "GEO TEST\n" + lines[0] + "\n" + lines[1]
	source := testCatalogSource(time.Date(2000, time.June, 28, 0, 0, 0, 0, time.UTC))
	records, err := ReadTLECatalog(strings.NewReader(content), source, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadTLECatalog: %v", err)
	}
	if records[0].PropagationSupported || records[0].PropagationLimitation == "" {
		t.Fatalf("unsupported orbit was not annotated: %+v", records[0])
	}
}

func TestReadTLECatalogRejectsInvalidAndDuplicateRecords(t *testing.T) {
	source := testCatalogSource(time.Date(2000, time.July, 1, 0, 0, 0, 0, time.UTC))
	tests := []struct {
		name        string
		content     string
		futureEpoch bool
	}{
		{name: "checksum failure", content: "VANGUARD 1\n" + norad5Line1[:68] + "0\n" + norad5Line2},
		{name: "line order", content: norad5Line2 + "\n" + norad5Line1},
		{name: "incomplete record", content: norad5Line1},
		{name: "duplicate catalog number", content: norad5Line1 + "\n" + norad5Line2 + "\n" + norad5Line1 + "\n" + norad5Line2},
		{name: "epoch after retrieval time", content: norad5Line1 + "\n" + norad5Line2, futureEpoch: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testSource := source
			if test.futureEpoch {
				testSource.RetrievedAtUTC = time.Date(2000, time.June, 27, 18, 0, 0, 0, time.UTC)
			}
			if _, err := ReadTLECatalog(strings.NewReader(test.content), testSource, 7*24*time.Hour); err == nil {
				t.Fatal("ReadTLECatalog succeeded for invalid input")
			}
		})
	}
}

func TestReadTLECatalogRequiresProvenanceAndFreshnessPolicy(t *testing.T) {
	source := testCatalogSource(time.Date(2000, time.July, 1, 0, 0, 0, 0, time.UTC))
	tests := []struct {
		name   string
		source CatalogSource
		maxAge time.Duration
	}{
		{name: "missing source", maxAge: 24 * time.Hour},
		{name: "missing snapshot ID", source: CatalogSource{Name: "fixture", RetrievedAtUTC: source.RetrievedAtUTC}, maxAge: 24 * time.Hour},
		{name: "missing retrieval time", source: CatalogSource{Name: "fixture", SnapshotID: "snapshot"}, maxAge: 24 * time.Hour},
		{name: "missing max age", source: source},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ReadTLECatalog(strings.NewReader(norad5Line1+"\n"+norad5Line2), test.source, test.maxAge); err == nil {
				t.Fatal("ReadTLECatalog succeeded without required metadata")
			}
		})
	}
}

func testCatalogSource(retrievedAt time.Time) CatalogSource {
	return CatalogSource{
		Name:           "Vallado verification fixture",
		URL:            "https://celestrak.org/publications/AIAA/2006-6753/",
		SnapshotID:     "vallado-sgp4-ver-test",
		RetrievedAtUTC: retrievedAt,
	}
}
