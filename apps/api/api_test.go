package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
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
