package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	handler, err := NewHandlerWithFixtureDir(filepath.Join("..", "..", "data", "fixtures"))
	if err != nil {
		t.Fatalf("load fixtures: %v", err)
	}
	return handler
}

func TestDefaultFixtureDiscovery(t *testing.T) {
	if _, err := NewHandler(); err != nil {
		t.Fatalf("discover repository fixtures: %v", err)
	}
}

func request(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

const vanguardTestOMM = `{
	"OBJECT_NAME":"VANGUARD 1",
	"OBJECT_ID":"1958-002B",
	"CENTER_NAME":"EARTH",
	"REF_FRAME":"TEME",
	"TIME_SYSTEM":"UTC",
	"MEAN_ELEMENT_THEORY":"SGP4",
	"EPOCH":"2000-06-27T18:50:19.733568",
	"MEAN_MOTION":10.82419157,
	"ECCENTRICITY":0.1859667,
	"INCLINATION":34.2682,
	"RA_OF_ASC_NODE":348.7242,
	"ARG_OF_PERICENTER":331.7664,
	"MEAN_ANOMALY":19.3264,
	"EPHEMERIS_TYPE":0,
	"CLASSIFICATION_TYPE":"U",
	"NORAD_CAT_ID":5,
	"ELEMENT_SET_NO":475,
	"REV_AT_EPOCH":41366,
	"BSTAR":0.000028098,
	"MEAN_MOTION_DOT":0.00000023,
	"MEAN_MOTION_DDOT":0
}`

func TestHealthCatalogAndFeaturedEndpoints(t *testing.T) {
	handler := testHandler(t)

	health := request(t, handler, http.MethodGet, "/api/health", "")
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"fixtureMode":true`) {
		t.Fatalf("health response = %d %s", health.Code, health.Body.String())
	}

	catalog := request(t, handler, http.MethodGet, "/api/catalog?query=lyra&limit=1", "")
	if catalog.Code != http.StatusOK {
		t.Fatalf("catalog status = %d: %s", catalog.Code, catalog.Body.String())
	}
	var page CatalogPage
	if err := json.Unmarshal(catalog.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].CatalogID != "sat-lyra-b" || page.NextCursor != nil {
		t.Fatalf("unexpected catalog page: %+v", page)
	}
	details := request(t, handler, http.MethodGet, "/api/catalog/sat-lyra-b", "")
	if details.Code != http.StatusOK || !strings.Contains(details.Body.String(), `"catalogId":"sat-lyra-b"`) {
		t.Fatalf("catalog detail response = %d %s", details.Code, details.Body.String())
	}

	featured := request(t, handler, http.MethodGet, "/api/demo/featured", "")
	if featured.Code != http.StatusOK || !strings.Contains(featured.Body.String(), `"featuredEncounterId":"enc-peregrine-lyra"`) {
		t.Fatalf("featured response = %d %s", featured.Code, featured.Body.String())
	}
}

func TestOMMImportEndpointValidatesWithoutChangingFixtureCatalog(t *testing.T) {
	handler := testHandler(t)
	omm := json.RawMessage(`[{
		"OBJECT_NAME":"VANGUARD 1",
		"OBJECT_ID":"1958-002B",
		"CENTER_NAME":"EARTH",
		"REF_FRAME":"TEME",
		"TIME_SYSTEM":"UTC",
		"MEAN_ELEMENT_THEORY":"SGP4",
		"EPOCH":"2000-06-27T18:50:19.733568",
		"MEAN_MOTION":10.82419157,
		"ECCENTRICITY":0.1859667,
		"INCLINATION":34.2682,
		"RA_OF_ASC_NODE":348.7242,
		"ARG_OF_PERICENTER":331.7664,
		"MEAN_ANOMALY":19.3264,
		"EPHEMERIS_TYPE":0,
		"CLASSIFICATION_TYPE":"U",
		"NORAD_CAT_ID":5,
		"ELEMENT_SET_NO":475,
		"REV_AT_EPOCH":41366,
		"BSTAR":0.000028098,
		"MEAN_MOTION_DOT":0.00000023,
		"MEAN_MOTION_DDOT":0,
		"COMMENT":"preserve source-specific fields"
	}]`)
	body, err := json.Marshal(OMMImportRequest{
		Source: OMMImportSource{
			Name:           "Vallado verification fixture",
			SnapshotID:     "vallado-omm-reference",
			RetrievedAtUTC: "2000-07-01T00:00:00Z",
		},
		MaxEpochAgeSeconds: 7 * 24 * 60 * 60,
		Records:            omm,
	})
	if err != nil {
		t.Fatal(err)
	}

	response := request(t, handler, http.MethodPost, "/api/catalog/omm", string(body))
	if response.Code != http.StatusOK {
		t.Fatalf("OMM import status = %d: %s", response.Code, response.Body.String())
	}
	var result OMMImportResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.SnapshotID != "vallado-omm-reference" || len(result.Records) != 1 ||
		result.Records[0].SatelliteNumber != 5 || !result.Records[0].PropagationSupported ||
		len(result.Caveats) == 0 ||
		!strings.Contains(string(result.Records[0].RawElement), `"COMMENT":"preserve source-specific fields"`) {
		t.Fatalf("unexpected OMM import result: %+v", result)
	}

	catalog := request(t, handler, http.MethodGet, "/api/catalog?limit=100", "")
	var page CatalogPage
	if err := json.Unmarshal(catalog.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 4 {
		t.Fatalf("fixture catalog changed after validation-only import: %d objects", len(page.Items))
	}
}

func TestOMMImportEndpointRejectsOversizedRequest(t *testing.T) {
	handler := testHandler(t)
	response := request(t, handler, http.MethodPost, "/api/catalog/omm", strings.Repeat(" ", (4<<20)+1))
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized OMM import status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestScreeningJobRunsAgainstCachedOMMSnapshot(t *testing.T) {
	handler := testHandler(t)
	records, err := os.ReadFile(filepath.Join("..", "..", "data", "reference", "screening-demo.omm.json"))
	if err != nil {
		t.Fatal(err)
	}
	importBody, err := json.Marshal(OMMImportRequest{
		Source: OMMImportSource{
			Name:           "Vallado verification fixture",
			SnapshotID:     "screening-api-test",
			RetrievedAtUTC: "2000-06-28T00:00:00Z",
		},
		MaxEpochAgeSeconds: 7 * 24 * 60 * 60,
		Records:            json.RawMessage(records),
	})
	if err != nil {
		t.Fatal(err)
	}
	importResponse := request(t, handler, http.MethodPost, "/api/catalog/omm", string(importBody))
	if importResponse.Code != http.StatusOK {
		t.Fatalf("OMM import status = %d: %s", importResponse.Code, importResponse.Body.String())
	}
	repeatedImport := request(t, handler, http.MethodPost, "/api/catalog/omm", string(importBody))
	if repeatedImport.Code != http.StatusOK {
		t.Fatalf("identical snapshot retry status = %d: %s", repeatedImport.Code, repeatedImport.Body.String())
	}
	conflictingBody := strings.Replace(string(importBody), `"MEAN_MOTION":10.82419157`, `"MEAN_MOTION":10.82419158`, 1)
	conflict := request(t, handler, http.MethodPost, "/api/catalog/omm", conflictingBody)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflicting snapshot ID status = %d, want %d: %s", conflict.Code, http.StatusConflict, conflict.Body.String())
	}

	createBody := `{
		"catalogSnapshotId":"screening-api-test",
		"startTimeUtc":"2000-06-27T19:07:00Z",
		"endTimeUtc":"2000-06-27T19:13:00Z",
		"thresholdMeters":10000,
		"coarseStepSeconds":60
	}`
	created := request(t, handler, http.MethodPost, "/api/screenings", createBody)
	if created.Code != http.StatusAccepted {
		t.Fatalf("screening create status = %d: %s", created.Code, created.Body.String())
	}
	var accepted CreateScreeningResponse
	if err := json.Unmarshal(created.Body.Bytes(), &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted.RunID == "" || accepted.Status != "queued" {
		t.Fatalf("unexpected screening response: %+v", accepted)
	}

	deadline := time.Now().Add(3 * time.Second)
	var status ScreeningStatus
	for time.Now().Before(deadline) {
		statusResponse := request(t, handler, http.MethodGet, "/api/screenings/"+accepted.RunID, "")
		if statusResponse.Code != http.StatusOK {
			t.Fatalf("screening status = %d: %s", statusResponse.Code, statusResponse.Body.String())
		}
		if err := json.Unmarshal(statusResponse.Body.Bytes(), &status); err != nil {
			t.Fatal(err)
		}
		if status.Status == "completed" || status.Status == "failed" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if status.Status != "completed" || status.Phase != "complete" || status.ProgressPercent != 100 {
		t.Fatalf("screening did not complete successfully: %+v", status)
	}

	results := request(t, handler, http.MethodGet, "/api/screenings/"+accepted.RunID+"/encounters?limit=1", "")
	if results.Code != http.StatusOK {
		t.Fatalf("screening results status = %d: %s", results.Code, results.Body.String())
	}
	var page EncounterPage
	if err := json.Unmarshal(results.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].RunID == nil || *page.Items[0].RunID != accepted.RunID ||
		page.Items[0].ObjectAID != "omm-norad-5" || page.Items[0].ObjectBID != "omm-norad-6" ||
		page.Items[0].MinimumSeparationM <= 0 || page.Items[0].RelativeSpeedMPS <= 0 ||
		len(page.Caveats) == 0 || page.NextCursor != nil {
		t.Fatalf("unexpected screening result page: %+v", page)
	}

	for i := 0; i < maxCachedOMMSnapshots; i++ {
		nextBody := strings.Replace(string(importBody), "screening-api-test", "replacement-"+string(rune('a'+i)), 1)
		if response := request(t, handler, http.MethodPost, "/api/catalog/omm", nextBody); response.Code != http.StatusOK {
			t.Fatalf("replacement import status = %d: %s", response.Code, response.Body.String())
		}
	}
	if response := request(t, handler, http.MethodPost, "/api/screenings", createBody); response.Code != http.StatusNotFound {
		t.Fatalf("expected original snapshot eviction, got %d: %s", response.Code, response.Body.String())
	}
	detailPath := "/api/screenings/" + accepted.RunID + "/encounters/" + page.Items[0].EncounterID
	detailResponse := request(t, handler, http.MethodGet, detailPath, "")
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("screened encounter detail status = %d: %s", detailResponse.Code, detailResponse.Body.String())
	}
	var detail ScreeningEncounterDetail
	if err := json.Unmarshal(detailResponse.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	trajectory := detail.Trajectory
	if trajectory.Frame.Frame != "TEME" || trajectory.Frame.VelocityUnit != "m/s" || len(trajectory.Samples) != 61 ||
		trajectory.Source.SnapshotID != "screening-api-test" || detail.Encounter.EncounterID != page.Items[0].EncounterID {
		t.Fatalf("unexpected screened trajectory metadata: %+v", trajectory.Frame)
	}
	tca := trajectory.Samples[30]
	if tca.AtUTC != detail.Encounter.TCAUTC || math.Abs(tca.SeparationMeters-detail.Encounter.MinimumSeparationM) > 1e-3 {
		t.Fatalf("center sample differs from reported TCA: %+v", tca)
	}
	for i, sample := range trajectory.Samples {
		at, err := time.Parse(time.RFC3339Nano, sample.AtUTC)
		center, _ := time.Parse(time.RFC3339Nano, tca.AtUTC)
		if err != nil || at.Sub(center) != time.Duration(i-30)*10*time.Second {
			t.Fatalf("unexpected timestamp for sample %d: %s", i, sample.AtUTC)
		}
	}
	if response := request(t, handler, http.MethodGet, detailPath+"-missing", ""); response.Code != http.StatusNotFound {
		t.Fatalf("unknown encounter returned %d", response.Code)
	}
	for _, input := range []string{`{}`, `{"alongTrackDeltaVMps":0.51}`, `{"alongTrackDeltaVMps":0,"extra":true}`} {
		if response := request(t, handler, http.MethodPost, detailPath+"/what-if", input); response.Code != http.StatusBadRequest {
			t.Fatalf("invalid what-if %s returned %d", input, response.Code)
		}
	}
	for _, input := range []string{`{"alongTrackDeltaVMps":0}`, `{"alongTrackDeltaVMps":0.5}`} {
		response := request(t, handler, http.MethodPost, detailPath+"/what-if", input)
		if response.Code != http.StatusOK {
			t.Fatalf("what-if returned %d: %s", response.Code, response.Body.String())
		}
		var preview ScreeningWhatIfResult
		if err := json.Unmarshal(response.Body.Bytes(), &preview); err != nil {
			t.Fatal(err)
		}
		if preview.ImpulseAtUTC != trajectory.StartUTC || preview.ObjectID != detail.Encounter.ObjectAID || len(preview.Samples) != len(trajectory.Samples) {
			t.Fatal("what-if grid or target does not match encounter detail")
		}
		for i, sample := range preview.Samples {
			baseline := trajectory.Samples[i]
			if sample.AtUTC != baseline.AtUTC || sample.ObjectB != baseline.ObjectB ||
				(preview.AlongTrackDeltaMPS == 0 && (sample.ObjectA != baseline.ObjectA || sample.SeparationMeters != baseline.SeparationMeters)) {
				t.Fatalf("what-if diverged from the expected baseline at sample %d", i)
			}
		}
		if preview.AlongTrackDeltaMPS != 0 && preview.Samples[30].SeparationMeters == tca.SeparationMeters {
			t.Fatal("nonzero impulse should change the fixture separation")
		}
	}
	if baseline := request(t, handler, http.MethodGet, detailPath, ""); baseline.Body.String() != detailResponse.Body.String() {
		t.Fatal("what-if modified the replay baseline")
	}
}

func TestScreeningJobRejectsMissingSnapshotAndOversizedWindow(t *testing.T) {
	handler := testHandler(t)
	validBody := `{
		"catalogSnapshotId":"not-imported",
		"startTimeUtc":"2000-06-27T19:07:00Z",
		"endTimeUtc":"2000-06-27T19:13:00Z",
		"thresholdMeters":10000,
		"coarseStepSeconds":60
	}`
	missing := request(t, handler, http.MethodPost, "/api/screenings", validBody)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing snapshot status = %d, want %d: %s", missing.Code, http.StatusNotFound, missing.Body.String())
	}

	oversized := request(t, handler, http.MethodPost, "/api/screenings", validBody+strings.Repeat(" ", (16<<10)+1))
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized screening request status = %d, want %d", oversized.Code, http.StatusRequestEntityTooLarge)
	}

	unbounded := request(t, handler, http.MethodPost, "/api/screenings", `{
		"catalogSnapshotId":"not-imported",
		"startTimeUtc":"2000-06-27T19:07:00Z",
		"endTimeUtc":"2000-06-27T20:07:00Z",
		"thresholdMeters":10000,
		"coarseStepSeconds":1
	}`)
	if unbounded.Code != http.StatusBadRequest {
		t.Fatalf("unbounded window status = %d, want %d: %s", unbounded.Code, http.StatusBadRequest, unbounded.Body.String())
	}
}

func TestEncounterTrajectoryAndManeuverFixtures(t *testing.T) {
	handler := testHandler(t)

	trajectory := request(t, handler, http.MethodGet, "/api/encounters/enc-peregrine-lyra/trajectory?samples=3", "")
	if trajectory.Code != http.StatusOK {
		t.Fatalf("trajectory status = %d: %s", trajectory.Code, trajectory.Body.String())
	}
	var result Trajectory
	if err := json.Unmarshal(trajectory.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Samples) != 3 || result.Samples[1].AtUTC != result.TCAUTC {
		t.Fatalf("trajectory does not include synchronized TCA sample: %+v", result.Samples)
	}
	if math.Abs(result.Samples[1].SeparationMeters-742) > 1e-6 {
		t.Fatalf("TCA separation = %f, want fixture minimum 742m", result.Samples[1].SeparationMeters)
	}
	details := request(t, handler, http.MethodGet, "/api/encounters/enc-peregrine-lyra", "")
	if details.Code != http.StatusOK || !strings.Contains(details.Body.String(), `"objectA"`) {
		t.Fatalf("encounter detail response = %d %s", details.Code, details.Body.String())
	}

	maneuver := request(t, handler, http.MethodPost, "/api/encounters/enc-peregrine-lyra/maneuvers",
		`{"maneuverTimeUtc":"2026-03-20T14:17:10Z","alongTrackDeltaVMps":0.1}`)
	if maneuver.Code != http.StatusOK {
		t.Fatalf("maneuver status = %d: %s", maneuver.Code, maneuver.Body.String())
	}
	var maneuverResult ManeuverResult
	if err := json.Unmarshal(maneuver.Body.Bytes(), &maneuverResult); err != nil {
		t.Fatal(err)
	}
	if maneuverResult.DatasetType != "synthetic" || len(maneuverResult.Limitations) == 0 ||
		maneuverResult.SimulatedMetrics.MinimumSeparationMeters == maneuverResult.OriginalMetrics.MinimumSeparationMeters {
		t.Fatalf("unexpected maneuver result: %+v", maneuverResult)
	}
}

func TestInvalidRequestsReturnContractErrors(t *testing.T) {
	handler := testHandler(t)
	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{http.MethodGet, "/api/encounters/enc-peregrine-lyra/trajectory?samples=1", "", http.StatusUnprocessableEntity},
		{http.MethodGet, "/api/encounters/missing/trajectory", "", http.StatusNotFound},
		{http.MethodPost, "/api/encounters/enc-peregrine-lyra/maneuvers", `{"maneuverTimeUtc":"2026-03-20T14:40:00Z","alongTrackDeltaVMps":0.1}`, http.StatusBadRequest},
		{http.MethodGet, "/api/catalog?limit=101", "", http.StatusBadRequest},
		{http.MethodPost, "/api/encounters/enc-peregrine-lyra/maneuvers", `{"maneuverTimeUtc":"2026-03-20T14:17:10Z"}`, http.StatusBadRequest},
		{http.MethodPost, "/api/catalog/omm", `{"source":{"retrievedAtUtc":"2000-07-01T00:00:00Z"},"maxEpochAgeSeconds":604800,"records":[]}`, http.StatusUnprocessableEntity},
		{http.MethodGet, "/api/unknown", "", http.StatusNotFound},
	}
	for _, test := range tests {
		recorder := request(t, handler, test.method, test.path, test.body)
		if recorder.Code != test.want {
			t.Errorf("%s %s status = %d, want %d: %s", test.method, test.path, recorder.Code, test.want, recorder.Body.String())
		}
		var response ErrorResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Code == "" || response.Details == nil {
			t.Errorf("%s %s did not return contract error object: %s", test.method, test.path, recorder.Body.String())
		}
	}
}
