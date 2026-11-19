package orbit

import (
	"strings"
	"testing"
	"time"
)

const vanguardOMM = `{
  "OBJECT_NAME": "VANGUARD 1",
  "OBJECT_ID": "1958-002B",
  "CENTER_NAME": "EARTH",
  "REF_FRAME": "TEME",
  "TIME_SYSTEM": "UTC",
  "MEAN_ELEMENT_THEORY": "SGP4",
  "EPOCH": "2000-06-27T18:50:19.733568",
  "MEAN_MOTION": 10.82419157,
  "ECCENTRICITY": 0.1859667,
  "INCLINATION": 34.2682,
  "RA_OF_ASC_NODE": 348.7242,
  "ARG_OF_PERICENTER": 331.7664,
  "MEAN_ANOMALY": 19.3264,
  "EPHEMERIS_TYPE": 0,
  "CLASSIFICATION_TYPE": "U",
  "NORAD_CAT_ID": 5,
  "ELEMENT_SET_NO": 475,
  "REV_AT_EPOCH": 41366,
  "BSTAR": 0.000028098,
  "MEAN_MOTION_DOT": 0.00000023,
  "MEAN_MOTION_DDOT": 0,
  "COMMENT": "preserve source-specific fields"
}`

func TestReadOMMCatalogParsesObjectAndArray(t *testing.T) {
	source := testCatalogSource(time.Date(2000, time.July, 1, 0, 0, 0, 0, time.UTC))
	for _, content := range []string{
		vanguardOMM,
		"[" + vanguardOMM + "]",
	} {
		records, err := ReadOMMCatalog(strings.NewReader(content), source, 7*24*time.Hour)
		if err != nil {
			t.Fatalf("ReadOMMCatalog: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("record count = %d, want 1", len(records))
		}
		record := records[0]
		if record.SatelliteNumber != 5 || record.Element.ObjectName != "VANGUARD 1" {
			t.Fatalf("unexpected identity: %+v", record)
		}
		if !record.PropagationSupported || record.DataQuality != "public_snapshot" {
			t.Fatalf("unexpected validation status: %+v", record)
		}
		if !strings.Contains(string(record.RawElement), `"COMMENT": "preserve source-specific fields"`) {
			t.Fatalf("raw OMM fields were not preserved: %s", record.RawElement)
		}
		if record.EpochUTC.Format(time.RFC3339Nano) != "2000-06-27T18:50:19.733568Z" {
			t.Fatalf("epoch = %s", record.EpochUTC.Format(time.RFC3339Nano))
		}
	}
}

func TestOMMPropagatorMatchesValladoNearEarthReference(t *testing.T) {
	records, err := ReadOMMCatalog(
		strings.NewReader(vanguardOMM),
		testCatalogSource(time.Date(2000, time.July, 1, 0, 0, 0, 0, time.UTC)),
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("ReadOMMCatalog: %v", err)
	}
	propagator, err := NewPropagatorFromOMM(records[0].Element)
	if err != nil {
		t.Fatalf("NewPropagatorFromOMM: %v", err)
	}
	got, err := propagator.Propagate(time.Date(2000, time.June, 27, 18, 50, 19, 733568000, time.UTC))
	if err != nil {
		t.Fatalf("Propagate: %v", err)
	}
	wantPosition := Vector3{X: 7022.46529266, Y: -1400.08296755, Z: 0.03995155}
	wantVelocity := Vector3{X: 1.893841015, Y: 6.405893759, Z: 4.534807250}
	if distance(got.PositionKM, wantPosition) > 1e-4 {
		t.Errorf("position = %+v, want %+v", got.PositionKM, wantPosition)
	}
	if distance(got.VelocityKMPS, wantVelocity) > 1e-6 {
		t.Errorf("velocity = %+v, want %+v", got.VelocityKMPS, wantVelocity)
	}
}

func TestReadOMMCatalogMarksStaleAndUnsupportedRecords(t *testing.T) {
	source := testCatalogSource(time.Date(2000, time.July, 10, 0, 0, 0, 0, time.UTC))
	stale, err := ReadOMMCatalog(strings.NewReader(vanguardOMM), source, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog: %v", err)
	}
	if stale[0].DataQuality != "stale" {
		t.Fatalf("DataQuality = %q, want stale", stale[0].DataQuality)
	}

	unsupportedOMM := strings.Replace(vanguardOMM, `"REF_FRAME": "TEME"`, `"REF_FRAME": "ITRF"`, 1)
	unsupported, err := ReadOMMCatalog(strings.NewReader(unsupportedOMM), source, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog unsupported frame: %v", err)
	}
	if unsupported[0].PropagationSupported || !strings.Contains(unsupported[0].PropagationLimitation, "REF_FRAME") {
		t.Fatalf("unsupported frame was not annotated: %+v", unsupported[0])
	}

	incompleteOMM := strings.Replace(vanguardOMM, `  "BSTAR": 0.000028098,`+"\n", "", 1)
	incomplete, err := ReadOMMCatalog(strings.NewReader(incompleteOMM), source, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog incomplete optional terms: %v", err)
	}
	if incomplete[0].PropagationSupported || len(incomplete[0].MissingFields) == 0 {
		t.Fatalf("missing propagation field was not annotated: %+v", incomplete[0])
	}
}

func TestReadOMMCatalogRejectsInvalidAndDuplicateRecords(t *testing.T) {
	source := testCatalogSource(time.Date(2000, time.July, 1, 0, 0, 0, 0, time.UTC))
	tests := []struct {
		name    string
		content string
		source  CatalogSource
	}{
		{name: "invalid eccentricity", content: strings.Replace(vanguardOMM, `"ECCENTRICITY": 0.1859667`, `"ECCENTRICITY": 1`, 1), source: source},
		{name: "missing epoch", content: strings.Replace(vanguardOMM, `  "EPOCH": "2000-06-27T18:50:19.733568",`+"\n", "", 1), source: source},
		{name: "future epoch", content: vanguardOMM, source: testCatalogSource(time.Date(2000, time.June, 27, 18, 0, 0, 0, time.UTC))},
		{name: "duplicate catalog number", content: "[" + vanguardOMM + "," + vanguardOMM + "]", source: source},
		{name: "invalid JSON", content: `{"OBJECT_NAME":`, source: source},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ReadOMMCatalog(strings.NewReader(test.content), test.source, 7*24*time.Hour); err == nil {
				t.Fatal("ReadOMMCatalog succeeded for invalid input")
			}
		})
	}
}

func TestReadOMMCatalogRequiresProvenanceAndFreshnessPolicy(t *testing.T) {
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
			if _, err := ReadOMMCatalog(strings.NewReader(vanguardOMM), test.source, test.maxAge); err == nil {
				t.Fatal("ReadOMMCatalog succeeded without required metadata")
			}
		})
	}
}
