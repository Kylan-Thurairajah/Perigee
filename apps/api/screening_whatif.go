package main

import (
	"errors"
	"math"
	"net/http"

	"github.com/perigee/perigee/apps/api/internal/orbit"
)

type ScreeningWhatIfRequest struct {
	AlongTrackDeltaVMPS *float64 `json:"alongTrackDeltaVMps"`
}

type ScreeningWhatIfResult struct {
	EncounterID        string             `json:"encounterId"`
	ObjectID           string             `json:"objectId"`
	ImpulseAtUTC       string             `json:"impulseAtUtc"`
	AlongTrackDeltaMPS float64            `json:"alongTrackDeltaVMps"`
	Model              string             `json:"model"`
	Samples            []TrajectorySample `json:"samples"`
	Caveats            []string           `json:"caveats"`
}

func (a *API) screeningWhatIf(w http.ResponseWriter, r *http.Request, runID, encounterID string) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var request ScreeningWhatIfRequest
	if err := decodeBody(r, &request); err != nil {
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "what-if request exceeds 1 KiB", nil)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_request", "provide a JSON object containing alongTrackDeltaVMps", nil)
		return
	}
	if request.AlongTrackDeltaVMPS == nil || math.Abs(*request.AlongTrackDeltaVMPS) > orbit.MaxReplayDeltaV {
		writeError(w, http.StatusBadRequest, "invalid_delta_v", "alongTrackDeltaVMps must be between -0.5 and 0.5 m/s", nil)
		return
	}
	detail, ok := a.screeningEncounterDetail(w, runID, encounterID)
	if !ok {
		return
	}
	trajectory := detail.Trajectory
	initial := trajectory.Samples[0].ObjectA
	delta := *request.AlongTrackDeltaVMPS
	offsets, err := orbit.ReplayImpulseOffsets(orbit.CartesianState{
		PositionKM:   orbit.Vector3{X: initial.PositionXKM, Y: initial.PositionYKM, Z: initial.PositionZKM},
		VelocityKMPS: orbit.Vector3{X: initial.VelocityXMPS / 1000, Y: initial.VelocityYMPS / 1000, Z: initial.VelocityZMPS / 1000},
	}, delta)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "what_if_failed", err.Error(), nil)
		return
	}
	for i, sample := range trajectory.Samples {
		offset := offsets[i]
		sample.ObjectA.PositionXKM += offset.PositionKM.X
		sample.ObjectA.PositionYKM += offset.PositionKM.Y
		sample.ObjectA.PositionZKM += offset.PositionKM.Z
		sample.ObjectA.VelocityXMPS += offset.VelocityKMPS.X * 1000
		sample.ObjectA.VelocityYMPS += offset.VelocityKMPS.Y * 1000
		sample.ObjectA.VelocityZMPS += offset.VelocityKMPS.Z * 1000
		sample.SeparationMeters = 1000 * math.Hypot(math.Hypot(
			sample.ObjectA.PositionXKM-sample.ObjectB.PositionXKM,
			sample.ObjectA.PositionYKM-sample.ObjectB.PositionYKM),
			sample.ObjectA.PositionZKM-sample.ObjectB.PositionZKM)
		trajectory.Samples[i] = sample
	}
	writeJSON(w, http.StatusOK, ScreeningWhatIfResult{
		EncounterID: encounterID, ObjectID: detail.Encounter.ObjectAID, ImpulseAtUTC: trajectory.StartUTC,
		AlongTrackDeltaMPS: delta, Model: "sgp4_baseline_plus_two_body_offset", Samples: trajectory.Samples,
		Caveats: []string{
			"The impulse is applied to object A parallel to its initial velocity at the replay start. The two-body offset uses central Earth gravity only and is added to the SGP4 baseline in TEME.",
			"Educational short-window approximation. Sampled separations do not provide collision probability or a maneuver recommendation.",
		},
	})
}
