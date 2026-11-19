package orbit

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	maxScreeningObjects = 100
	maxScreeningSamples = 1441
)

type ScreeningConfig struct {
	StartUTC          time.Time
	EndUTC            time.Time
	ThresholdMeters   float64
	CoarseStepSeconds int
}

type ScreenedConjunction struct {
	ObjectANORADID          int       `json:"objectANoradId"`
	ObjectAName             string    `json:"objectAName"`
	ObjectBNORADID          int       `json:"objectBNoradId"`
	ObjectBName             string    `json:"objectBName"`
	TCAUTC                  time.Time `json:"tcaUtc"`
	MinimumSeparationMeters float64   `json:"minimumSeparationMeters"`
	RelativeSpeedMetersPerS float64   `json:"relativeSpeedMetersPerSecond"`
}

type ScreeningResult struct {
	CatalogSnapshotID string                `json:"catalogSnapshotId"`
	StartUTC          time.Time             `json:"startUtc"`
	EndUTC            time.Time             `json:"endUtc"`
	ThresholdMeters   float64               `json:"thresholdMeters"`
	CoarseStepSeconds int                   `json:"coarseStepSeconds"`
	SampleCount       int                   `json:"sampleCount"`
	ObjectsConsidered int                   `json:"objectsConsidered"`
	PairsEvaluated    int                   `json:"pairsEvaluated"`
	Encounters        []ScreenedConjunction `json:"encounters"`
	Caveats           []string              `json:"caveats"`
}

type sampledObject struct {
	record     OMMCatalogRecord
	propagator *Propagator
	states     []TEMEState
}

func ScreenOMMCatalog(records []OMMCatalogRecord, config ScreeningConfig) (ScreeningResult, error) {
	if len(records) < 2 {
		return ScreeningResult{}, fmt.Errorf("at least two catalog records are required")
	}
	if len(records) > maxScreeningObjects {
		return ScreeningResult{}, fmt.Errorf("screening supports at most %d objects per run", maxScreeningObjects)
	}
	sampleCount, err := screeningSampleCount(config)
	if err != nil {
		return ScreeningResult{}, err
	}
	start := config.StartUTC.UTC()
	end := config.EndUTC.UTC()
	step := time.Duration(config.CoarseStepSeconds) * time.Second

	ordered := append([]OMMCatalogRecord(nil), records...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].SatelliteNumber < ordered[j].SatelliteNumber
	})
	snapshotID := strings.TrimSpace(ordered[0].Source.SnapshotID)
	if snapshotID == "" {
		return ScreeningResult{}, fmt.Errorf("catalog snapshot ID is required")
	}
	objects := make([]sampledObject, len(ordered))
	for index, record := range ordered {
		if index > 0 && record.SatelliteNumber == ordered[index-1].SatelliteNumber {
			return ScreeningResult{}, fmt.Errorf("duplicate catalog number %d", record.SatelliteNumber)
		}
		if record.DataQuality == "stale" {
			return ScreeningResult{}, fmt.Errorf("catalog number %d has stale orbital elements", record.SatelliteNumber)
		}
		if !record.PropagationSupported {
			return ScreeningResult{}, fmt.Errorf("catalog number %d is not propagation-supported: %s", record.SatelliteNumber, record.PropagationLimitation)
		}
		if strings.TrimSpace(record.Source.SnapshotID) != snapshotID {
			return ScreeningResult{}, fmt.Errorf("catalog records must share one snapshot ID")
		}
		propagator, err := NewPropagatorFromOMM(record.Element)
		if err != nil {
			return ScreeningResult{}, fmt.Errorf("initialize catalog number %d: %w", record.SatelliteNumber, err)
		}
		objects[index] = sampledObject{record: record, propagator: propagator}
	}

	times := make([]time.Time, 0, int(sampleCount))
	for at := start; at.Before(end); at = at.Add(step) {
		times = append(times, at)
	}
	times = append(times, end)

	for objectIndex := range objects {
		object := &objects[objectIndex]
		object.states = make([]TEMEState, len(times))
		for sampleIndex, at := range times {
			state, err := object.propagator.Propagate(at)
			if err != nil {
				return ScreeningResult{}, fmt.Errorf("propagate catalog number %d at %s: %w", object.record.SatelliteNumber, at.Format(time.RFC3339Nano), err)
			}
			object.states[sampleIndex] = state
		}
	}

	result := ScreeningResult{
		CatalogSnapshotID: snapshotID,
		StartUTC:          start,
		EndUTC:            end,
		ThresholdMeters:   config.ThresholdMeters,
		CoarseStepSeconds: config.CoarseStepSeconds,
		SampleCount:       len(times),
		ObjectsConsidered: len(objects),
		PairsEvaluated:    len(objects) * (len(objects) - 1) / 2,
		Encounters:        make([]ScreenedConjunction, 0),
		Caveats: []string{
			"Piecewise-linear coarse segments select one candidate interval and one minimum per object pair; a coarse step that is too large may miss close approaches.",
			"Minimum separation is a miss distance, not collision probability. No uncertainty model is applied.",
			"States are SGP4 TEME outputs; no frame conversion is performed.",
			"Each run is limited to 100 objects and 1441 UTC samples.",
		},
	}

	thresholdKM := config.ThresholdMeters / 1000
	for first := 0; first < len(objects); first++ {
		for second := first + 1; second < len(objects); second++ {
			encounters, err := appendIfWithinThreshold(result.Encounters, objects[first], objects[second], times, thresholdKM)
			if err != nil {
				return ScreeningResult{}, err
			}
			result.Encounters = encounters
		}
	}
	sort.Slice(result.Encounters, func(i, j int) bool {
		if result.Encounters[i].TCAUTC.Equal(result.Encounters[j].TCAUTC) {
			if result.Encounters[i].ObjectANORADID == result.Encounters[j].ObjectANORADID {
				return result.Encounters[i].ObjectBNORADID < result.Encounters[j].ObjectBNORADID
			}
			return result.Encounters[i].ObjectANORADID < result.Encounters[j].ObjectANORADID
		}
		return result.Encounters[i].TCAUTC.Before(result.Encounters[j].TCAUTC)
	})
	return result, nil
}

func ValidateScreeningConfig(config ScreeningConfig) error {
	_, err := screeningSampleCount(config)
	return err
}

func screeningSampleCount(config ScreeningConfig) (int64, error) {
	if math.IsNaN(config.ThresholdMeters) || math.IsInf(config.ThresholdMeters, 0) || config.ThresholdMeters <= 0 {
		return 0, fmt.Errorf("threshold must be a positive finite distance in meters")
	}
	if config.CoarseStepSeconds < 1 || config.CoarseStepSeconds > 3600 {
		return 0, fmt.Errorf("coarse step must be between 1 and 3600 seconds")
	}
	if config.StartUTC.IsZero() || config.EndUTC.IsZero() || !config.EndUTC.After(config.StartUTC) {
		return 0, fmt.Errorf("screening end time must be after a non-zero start time")
	}
	step := time.Duration(config.CoarseStepSeconds) * time.Second
	duration := config.EndUTC.Sub(config.StartUTC)
	intervalCount := int64(duration / step)
	if duration%step != 0 {
		intervalCount++
	}
	sampleCount := intervalCount + 1
	if sampleCount > maxScreeningSamples {
		return 0, fmt.Errorf("screening requires %d samples; maximum is %d", sampleCount, maxScreeningSamples)
	}
	return sampleCount, nil
}

func appendIfWithinThreshold(
	encounters []ScreenedConjunction,
	first, second sampledObject,
	times []time.Time,
	thresholdKM float64,
) ([]ScreenedConjunction, error) {
	bestInterval := -1
	bestFraction := 0.0
	bestSegmentDistance := math.Inf(1)
	for index := 0; index < len(times)-1; index++ {
		startRelative := subtract(first.states[index].PositionKM, second.states[index].PositionKM)
		endRelative := subtract(first.states[index+1].PositionKM, second.states[index+1].PositionKM)
		fraction, distanceKM := closestPointOnSegment(startRelative, endRelative)
		if distanceKM < bestSegmentDistance {
			bestInterval = index
			bestFraction = fraction
			bestSegmentDistance = distanceKM
		}
	}
	if bestInterval < 0 || bestSegmentDistance > thresholdKM {
		return encounters, nil
	}

	start := times[bestInterval]
	end := times[bestInterval+1]
	tca := interpolateTime(start, end, bestFraction)
	bestFirst, err := first.propagator.Propagate(tca)
	if err != nil {
		return nil, fmt.Errorf("refine catalog pair %d/%d: %w", first.record.SatelliteNumber, second.record.SatelliteNumber, err)
	}
	bestSecond, err := second.propagator.Propagate(tca)
	if err != nil {
		return nil, fmt.Errorf("refine catalog pair %d/%d: %w", first.record.SatelliteNumber, second.record.SatelliteNumber, err)
	}
	bestDistance := vectorNorm(subtract(bestFirst.PositionKM, bestSecond.PositionKM))

	left := start
	right := end
	const goldenRatio = 0.6180339887498949
	leftProbe := right.Add(-time.Duration(float64(right.Sub(left)) * goldenRatio))
	rightProbe := left.Add(time.Duration(float64(right.Sub(left)) * goldenRatio))
	leftStateA, leftStateB, leftDistance, err := propagatePair(first, second, leftProbe)
	if err != nil {
		return nil, fmt.Errorf("refine catalog pair %d/%d: %w", first.record.SatelliteNumber, second.record.SatelliteNumber, err)
	}
	rightStateA, rightStateB, rightDistance, err := propagatePair(first, second, rightProbe)
	if err != nil {
		return nil, fmt.Errorf("refine catalog pair %d/%d: %w", first.record.SatelliteNumber, second.record.SatelliteNumber, err)
	}
	for iteration := 0; iteration < 32; iteration++ {
		if leftDistance < bestDistance {
			tca, bestFirst, bestSecond, bestDistance = leftProbe, leftStateA, leftStateB, leftDistance
		}
		if rightDistance < bestDistance {
			tca, bestFirst, bestSecond, bestDistance = rightProbe, rightStateA, rightStateB, rightDistance
		}
		if leftDistance <= rightDistance {
			right = rightProbe
			rightProbe, rightStateA, rightStateB, rightDistance = leftProbe, leftStateA, leftStateB, leftDistance
			leftProbe = right.Add(-time.Duration(float64(right.Sub(left)) * goldenRatio))
			leftStateA, leftStateB, leftDistance, err = propagatePair(first, second, leftProbe)
		} else {
			left = leftProbe
			leftProbe, leftStateA, leftStateB, leftDistance = rightProbe, rightStateA, rightStateB, rightDistance
			rightProbe = left.Add(time.Duration(float64(right.Sub(left)) * goldenRatio))
			rightStateA, rightStateB, rightDistance, err = propagatePair(first, second, rightProbe)
		}
		if err != nil {
			return nil, fmt.Errorf("refine catalog pair %d/%d: %w", first.record.SatelliteNumber, second.record.SatelliteNumber, err)
		}
	}
	if bestDistance > thresholdKM {
		return encounters, nil
	}

	relativeVelocity := subtract(bestFirst.VelocityKMPS, bestSecond.VelocityKMPS)
	encounters = append(encounters, ScreenedConjunction{
		ObjectANORADID:          first.record.SatelliteNumber,
		ObjectAName:             first.record.Element.ObjectName,
		ObjectBNORADID:          second.record.SatelliteNumber,
		ObjectBName:             second.record.Element.ObjectName,
		TCAUTC:                  tca.UTC(),
		MinimumSeparationMeters: bestDistance * 1000,
		RelativeSpeedMetersPerS: vectorNorm(relativeVelocity) * 1000,
	})
	return encounters, nil
}

func propagatePair(first, second sampledObject, at time.Time) (TEMEState, TEMEState, float64, error) {
	stateA, err := first.propagator.Propagate(at)
	if err != nil {
		return TEMEState{}, TEMEState{}, 0, err
	}
	stateB, err := second.propagator.Propagate(at)
	if err != nil {
		return TEMEState{}, TEMEState{}, 0, err
	}
	return stateA, stateB, vectorNorm(subtract(stateA.PositionKM, stateB.PositionKM)), nil
}

func closestPointOnSegment(start, end Vector3) (float64, float64) {
	delta := subtract(end, start)
	lengthSquared := dot(delta, delta)
	if lengthSquared == 0 {
		return 0, vectorNorm(start)
	}
	fraction := -dot(start, delta) / lengthSquared
	if fraction < 0 {
		fraction = 0
	} else if fraction > 1 {
		fraction = 1
	}
	return fraction, vectorNorm(add(start, scale(delta, fraction)))
}

func interpolateTime(start, end time.Time, fraction float64) time.Time {
	return start.Add(time.Duration(float64(end.Sub(start)) * fraction))
}

func subtract(first, second Vector3) Vector3 {
	return Vector3{X: first.X - second.X, Y: first.Y - second.Y, Z: first.Z - second.Z}
}

func add(first, second Vector3) Vector3 {
	return Vector3{X: first.X + second.X, Y: first.Y + second.Y, Z: first.Z + second.Z}
}

func scale(value Vector3, factor float64) Vector3 {
	return Vector3{X: value.X * factor, Y: value.Y * factor, Z: value.Z * factor}
}

func dot(first, second Vector3) float64 {
	return first.X*second.X + first.Y*second.Y + first.Z*second.Z
}

func vectorNorm(value Vector3) float64 {
	return math.Sqrt(dot(value, value))
}
