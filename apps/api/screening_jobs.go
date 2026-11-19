package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/perigee/perigee/apps/api/internal/orbit"
)

const (
	maxCachedOMMSnapshots = 4
	maxScreeningRuns      = 100
	screeningWorkerCount  = 2
	screeningQueueSize    = 8
)

type CreateScreeningRequest struct {
	CatalogSnapshotID string  `json:"catalogSnapshotId"`
	StartTimeUTC      string  `json:"startTimeUtc"`
	EndTimeUTC        string  `json:"endTimeUtc"`
	ThresholdMeters   float64 `json:"thresholdMeters"`
	CoarseStepSeconds int     `json:"coarseStepSeconds"`
}

type CreateScreeningResponse struct {
	RunID  string `json:"runId"`
	Status string `json:"status"`
}

type ScreeningStatus struct {
	RunID           string         `json:"runId"`
	Status          string         `json:"status"`
	Phase           string         `json:"phase"`
	ProgressPercent int            `json:"progressPercent"`
	CreatedAtUTC    time.Time      `json:"createdAtUtc"`
	UpdatedAtUTC    time.Time      `json:"updatedAtUtc"`
	StartedAtUTC    *time.Time     `json:"startedAtUtc,omitempty"`
	CompletedAtUTC  *time.Time     `json:"completedAtUtc,omitempty"`
	Error           *ErrorResponse `json:"error,omitempty"`
}

type EncounterPage struct {
	Items      []EncounterSummary `json:"items"`
	NextCursor *string            `json:"nextCursor"`
	Caveats    []string           `json:"caveats"`
}

type ommSnapshot struct {
	source     orbit.CatalogSource
	records    []orbit.OMMCatalogRecord
	lastUsedAt time.Time
}

type screeningRun struct {
	status   ScreeningStatus
	result   *orbit.ScreeningResult
	elements map[int]orbit.OMMCatalogRecord
	source   orbit.CatalogSource
}

type screeningTask struct {
	runID   string
	records []orbit.OMMCatalogRecord
	config  orbit.ScreeningConfig
}

func (a *API) registerOMMSnapshot(source orbit.CatalogSource, records []orbit.OMMCatalogRecord) error {
	now := time.Now().UTC()
	a.mu.Lock()
	defer a.mu.Unlock()

	if existing, ok := a.ommSnapshots[source.SnapshotID]; ok {
		if !reflect.DeepEqual(existing.source, source) || !reflect.DeepEqual(existing.records, records) {
			return fmt.Errorf("snapshot ID %q is already registered with different contents", source.SnapshotID)
		}
		existing.lastUsedAt = now
		a.ommSnapshots[source.SnapshotID] = existing
		return nil
	}

	if len(a.ommSnapshots) >= maxCachedOMMSnapshots {
		oldestID := ""
		var oldestTime time.Time
		for snapshotID, snapshot := range a.ommSnapshots {
			if oldestID == "" || snapshot.lastUsedAt.Before(oldestTime) ||
				(snapshot.lastUsedAt.Equal(oldestTime) && snapshotID < oldestID) {
				oldestID = snapshotID
				oldestTime = snapshot.lastUsedAt
			}
		}
		delete(a.ommSnapshots, oldestID)
	}
	a.ommSnapshots[source.SnapshotID] = ommSnapshot{
		source: source, records: append([]orbit.OMMCatalogRecord(nil), records...), lastUsedAt: now,
	}
	return nil
}

func (a *API) createScreening(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request CreateScreeningRequest
	if err := decodeBody(r, &request); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "screening request exceeds the 16 KiB limit", nil)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON with only documented fields", map[string]any{"reason": err.Error()})
		return
	}
	request.CatalogSnapshotID = strings.TrimSpace(request.CatalogSnapshotID)
	if request.CatalogSnapshotID == "" {
		writeError(w, http.StatusBadRequest, "invalid_catalog_snapshot", "catalogSnapshotId is required", nil)
		return
	}
	start, err := parseUTC(request.StartTimeUTC)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_start_time", "startTimeUtc must be an RFC3339 UTC timestamp ending in Z", nil)
		return
	}
	end, err := parseUTC(request.EndTimeUTC)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_end_time", "endTimeUtc must be an RFC3339 UTC timestamp ending in Z", nil)
		return
	}
	config := orbit.ScreeningConfig{
		StartUTC: start, EndUTC: end, ThresholdMeters: request.ThresholdMeters,
		CoarseStepSeconds: request.CoarseStepSeconds,
	}
	if err := orbit.ValidateScreeningConfig(config); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_screening_config", err.Error(), nil)
		return
	}

	runID, err := newScreeningRunID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "screening_id_failed", "could not allocate a screening run ID", nil)
		return
	}
	now := time.Now().UTC()

	a.mu.Lock()
	defer a.mu.Unlock()
	snapshot, ok := a.ommSnapshots[request.CatalogSnapshotID]
	if !ok {
		writeError(w, http.StatusNotFound, "catalog_snapshot_not_found", "catalog snapshot is not in the in-memory cache", map[string]any{"catalogSnapshotId": request.CatalogSnapshotID})
		return
	}
	if len(snapshot.records) < 2 {
		writeError(w, http.StatusUnprocessableEntity, "catalog_not_screenable", "screening requires at least two catalog records", nil)
		return
	}
	if len(snapshot.records) > 100 {
		writeError(w, http.StatusUnprocessableEntity, "catalog_not_screenable", "screening supports at most 100 catalog records", map[string]any{"recordCount": len(snapshot.records), "maximum": 100})
		return
	}
	for _, record := range snapshot.records {
		if record.DataQuality == "stale" || !record.PropagationSupported {
			reason := record.PropagationLimitation
			if record.DataQuality == "stale" {
				reason = "orbital elements are stale"
			}
			writeError(w, http.StatusUnprocessableEntity, "catalog_not_screenable", "every catalog record must be fresh and supported by the selected SGP4 adapter", map[string]any{
				"noradCatId": record.SatelliteNumber, "reason": reason,
			})
			return
		}
	}

	if len(a.screeningQueue) == cap(a.screeningQueue) {
		writeError(w, http.StatusServiceUnavailable, "screening_queue_full", "screening queue is full; retry after a run completes", nil)
		return
	}
	if !a.makeScreeningRunRoom() {
		writeError(w, http.StatusServiceUnavailable, "screening_capacity_reached", "the retained screening run limit has been reached while all runs are active", nil)
		return
	}
	elements := make(map[int]orbit.OMMCatalogRecord, len(snapshot.records))
	for _, record := range snapshot.records {
		record.RawElement = nil
		elements[record.SatelliteNumber] = record
	}
	run := &screeningRun{
		status: ScreeningStatus{
			RunID: runID, Status: "queued", Phase: "queued",
			CreatedAtUTC: now, UpdatedAtUTC: now,
		},
		elements: elements, source: snapshot.source,
	}
	a.screenings[runID] = run
	snapshot.lastUsedAt = now
	a.ommSnapshots[request.CatalogSnapshotID] = snapshot

	task := screeningTask{
		runID: runID, records: append([]orbit.OMMCatalogRecord(nil), snapshot.records...), config: config,
	}
	select {
	case a.screeningQueue <- task:
		a.startScreeningWorkers()
		writeJSON(w, http.StatusAccepted, CreateScreeningResponse{RunID: runID, Status: "queued"})
	default:
		delete(a.screenings, runID)
		writeError(w, http.StatusServiceUnavailable, "screening_queue_full", "screening queue is full; retry after a run completes", nil)
	}
}

func (a *API) makeScreeningRunRoom() bool {
	if len(a.screenings) < maxScreeningRuns {
		return true
	}
	oldestID := ""
	var oldestTime time.Time
	for runID, run := range a.screenings {
		if run.status.Status != "completed" && run.status.Status != "failed" && run.status.Status != "cancelled" {
			continue
		}
		if oldestID == "" || run.status.CreatedAtUTC.Before(oldestTime) ||
			(run.status.CreatedAtUTC.Equal(oldestTime) && runID < oldestID) {
			oldestID = runID
			oldestTime = run.status.CreatedAtUTC
		}
	}
	if oldestID == "" {
		return false
	}
	delete(a.screenings, oldestID)
	return true
}

func (a *API) startScreeningWorkers() {
	a.screeningWorkersOnce.Do(func() {
		for range screeningWorkerCount {
			go a.screeningWorker()
		}
	})
}

func (a *API) screeningWorker() {
	for task := range a.screeningQueue {
		now := time.Now().UTC()
		a.mu.Lock()
		run := a.screenings[task.runID]
		if run != nil {
			run.status.Status = "running"
			run.status.Phase = "propagating"
			run.status.ProgressPercent = 10
			run.status.StartedAtUTC = &now
			run.status.UpdatedAtUTC = now
		}
		a.mu.Unlock()
		if run == nil {
			continue
		}

		result, err := orbit.ScreenOMMCatalog(task.records, task.config)
		completed := time.Now().UTC()
		a.mu.Lock()
		run = a.screenings[task.runID]
		if run != nil {
			run.status.UpdatedAtUTC = completed
			run.status.CompletedAtUTC = &completed
			run.status.ProgressPercent = 100
			if err != nil {
				run.status.Status = "failed"
				run.status.Phase = "failed"
				run.status.Error = &ErrorResponse{
					Code: "screening_failed", Message: "screening run failed",
					Details: map[string]any{"reason": err.Error()},
				}
			} else {
				run.status.Status = "completed"
				run.status.Phase = "complete"
				run.result = &result
			}
		}
		a.mu.Unlock()
	}
}

func (a *API) getScreeningStatus(w http.ResponseWriter, runID string) {
	a.mu.Lock()
	run, ok := a.screenings[runID]
	if ok {
		run.status.UpdatedAtUTC = run.status.UpdatedAtUTC.UTC()
	}
	var status ScreeningStatus
	if ok {
		status = run.status
	}
	a.mu.Unlock()
	if !ok {
		writeError(w, http.StatusNotFound, "screening_not_found", "screening run was not found", map[string]any{"runId": runID})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (a *API) listScreeningEncounters(w http.ResponseWriter, r *http.Request, runID string) {
	run, result := a.completedScreening(w, runID)
	if result == nil {
		return
	}

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

	items := make([]EncounterSummary, 0, len(result.Encounters))
	for _, encounter := range result.Encounters {
		items = append(items, run.encounterSummary(result.CatalogSnapshotID, encounter))
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
	writeJSON(w, http.StatusOK, EncounterPage{
		Items: pageItems, NextCursor: next, Caveats: append([]string(nil), result.Caveats...),
	})
}

func (a *API) completedScreening(w http.ResponseWriter, runID string) (*screeningRun, *orbit.ScreeningResult) {
	a.mu.Lock()
	run, ok := a.screenings[runID]
	var status string
	var result *orbit.ScreeningResult
	if ok {
		status = run.status.Status
		result = run.result
	}
	a.mu.Unlock()
	if !ok {
		writeError(w, http.StatusNotFound, "screening_not_found", "screening run was not found", map[string]any{"runId": runID})
		return nil, nil
	}
	if status != "completed" || result == nil {
		writeError(w, http.StatusConflict, "screening_not_complete", "screening encounters are available only after a successful run", map[string]any{"status": status})
		return nil, nil
	}

	return run, result
}

func (run *screeningRun) encounterSummary(snapshotID string, encounter orbit.ScreenedConjunction) EncounterSummary {
	runID := run.status.RunID
	sourceEpoch := run.elements[encounter.ObjectANORADID].EpochUTC
	if epochB := run.elements[encounter.ObjectBNORADID].EpochUTC; epochB.Before(sourceEpoch) {
		sourceEpoch = epochB
	}
	return EncounterSummary{
		EncounterID:       fmt.Sprintf("enc-%s-%d-%d", runID, encounter.ObjectANORADID, encounter.ObjectBNORADID),
		ObjectAID:         fmt.Sprintf("omm-norad-%d", encounter.ObjectANORADID),
		ObjectBID:         fmt.Sprintf("omm-norad-%d", encounter.ObjectBNORADID),
		CatalogSnapshotID: snapshotID, RunID: &runID,
		TCAUTC:             encounter.TCAUTC.UTC().Format(time.RFC3339Nano),
		MinimumSeparationM: encounter.MinimumSeparationMeters,
		RelativeSpeedMPS:   encounter.RelativeSpeedMetersPerS,
		SourceEpochUTC:     sourceEpoch.UTC().Format(time.RFC3339Nano),
		DatasetType:        "public_snapshot", DataQuality: "public_snapshot",
	}
}

type ScreeningEncounterDetail struct {
	Encounter   EncounterSummary `json:"encounter"`
	ObjectAName string           `json:"objectAName"`
	ObjectBName string           `json:"objectBName"`
	Trajectory  Trajectory       `json:"trajectory"`
	Caveats     []string         `json:"caveats"`
}

func (a *API) getScreeningEncounter(w http.ResponseWriter, runID, encounterID string) {
	detail, ok := a.screeningEncounterDetail(w, runID, encounterID)
	if ok {
		writeJSON(w, http.StatusOK, detail)
	}
}

func (a *API) screeningEncounterDetail(w http.ResponseWriter, runID, encounterID string) (ScreeningEncounterDetail, bool) {
	run, result := a.completedScreening(w, runID)
	if result == nil {
		return ScreeningEncounterDetail{}, false
	}
	for _, encounter := range result.Encounters {
		summary := run.encounterSummary(result.CatalogSnapshotID, encounter)
		if summary.EncounterID != encounterID {
			continue
		}
		trajectory, err := run.encounterTrajectory(summary, encounter)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "screening_trajectory_failed", "could not propagate the encounter replay window", map[string]any{"reason": err.Error()})
			return ScreeningEncounterDetail{}, false
		}
		return ScreeningEncounterDetail{
			Encounter: summary, ObjectAName: encounter.ObjectAName, ObjectBName: encounter.ObjectBName,
			Trajectory: trajectory, Caveats: result.Caveats,
		}, true
	}
	writeError(w, http.StatusNotFound, "screening_encounter_not_found", "encounter was not found in the screening run", nil)
	return ScreeningEncounterDetail{}, false
}

func (run *screeningRun) encounterTrajectory(summary EncounterSummary, encounter orbit.ScreenedConjunction) (Trajectory, error) {
	first, err := orbit.NewPropagatorFromOMM(run.elements[encounter.ObjectANORADID].Element)
	if err != nil {
		return Trajectory{}, err
	}
	second, err := orbit.NewPropagatorFromOMM(run.elements[encounter.ObjectBNORADID].Element)
	if err != nil {
		return Trajectory{}, err
	}
	const halfWindow = orbit.ReplayStep * (orbit.ReplaySampleCount / 2)
	start, end := encounter.TCAUTC.Add(-halfWindow), encounter.TCAUTC.Add(halfWindow)
	samples := make([]TrajectorySample, 0, orbit.ReplaySampleCount)
	for at := start; !at.After(end); at = at.Add(orbit.ReplayStep) {
		stateA, err := first.Propagate(at)
		if err != nil {
			return Trajectory{}, err
		}
		stateB, err := second.Propagate(at)
		if err != nil {
			return Trajectory{}, err
		}
		deltaX := stateA.PositionKM.X - stateB.PositionKM.X
		deltaY := stateA.PositionKM.Y - stateB.PositionKM.Y
		deltaZ := stateA.PositionKM.Z - stateB.PositionKM.Z
		samples = append(samples, TrajectorySample{
			AtUTC: at.UTC().Format(time.RFC3339Nano), ObjectA: screeningObjectState(stateA), ObjectB: screeningObjectState(stateB),
			SeparationMeters: 1000 * math.Hypot(math.Hypot(deltaX, deltaY), deltaZ),
		})
	}
	var sourceURL *string
	if run.source.URL != "" {
		sourceURL = &run.source.URL
	}
	return Trajectory{
		EncounterID: summary.EncounterID, TCAUTC: summary.TCAUTC, Samples: samples,
		StartUTC: start.UTC().Format(time.RFC3339Nano), EndUTC: end.UTC().Format(time.RFC3339Nano),
		Frame: FrameMetadata{Frame: "TEME", PositionUnit: "km", VelocityUnit: "m/s", DistanceUnit: "m", FrameCaveat: "Earth-centered TEME axes. The displayed Earth is illustrative."},
		Source: SourceMetadata{SourceName: run.source.Name, SourceURL: sourceURL, SnapshotID: run.source.SnapshotID,
			RetrievedAtUTC: run.source.RetrievedAtUTC.UTC().Format(time.RFC3339Nano), DatasetType: summary.DatasetType,
			Caveat: "SGP4 replay extends five minutes either side of the reported TCA and may extend past the screening window. The reported TCA is limited to the screening window."},
	}, nil
}

func screeningObjectState(state orbit.TEMEState) ObjectState {
	return ObjectState{
		PositionXKM: state.PositionKM.X, PositionYKM: state.PositionKM.Y, PositionZKM: state.PositionKM.Z,
		VelocityXMPS: 1000 * state.VelocityKMPS.X, VelocityYMPS: 1000 * state.VelocityKMPS.Y, VelocityZMPS: 1000 * state.VelocityKMPS.Z,
	}
}

func newScreeningRunID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func (a *API) serveScreeningRun(w http.ResponseWriter, r *http.Request) {
	runID, resource, encounterID, ok := parseScreeningRunPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "route not found", nil)
		return
	}
	method := http.MethodGet
	if resource == "what-if" {
		method = http.MethodPost
	}
	if r.Method != method {
		w.Header().Set("Allow", method)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported", nil)
		return
	}
	if resource == "what-if" {
		a.screeningWhatIf(w, r, runID, encounterID)
	} else if encounterID != "" {
		a.getScreeningEncounter(w, runID, encounterID)
	} else if resource == "encounters" {
		a.listScreeningEncounters(w, r, runID)
	} else {
		a.getScreeningStatus(w, runID)
	}
}

func parseScreeningRunPath(path string) (string, string, string, bool) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(path, "/api/screenings/"), "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", "", true
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "encounters" {
		return parts[0], parts[1], "", true
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] == "encounters" && parts[2] != "" {
		return parts[0], parts[1], parts[2], true
	}
	if len(parts) == 4 && parts[0] != "" && parts[1] == "encounters" && parts[2] != "" && parts[3] == "what-if" {
		return parts[0], parts[3], parts[2], true
	}
	return "", "", "", false
}
