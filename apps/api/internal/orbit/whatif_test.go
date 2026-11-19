package orbit

import (
	"math"
	"testing"
)

func TestTwoBodyStepFollowsCircularOrbit(t *testing.T) {
	const radius = 7000.0
	speed := math.Sqrt(earthMuKM3PerS2 / radius)
	state := CartesianState{PositionKM: Vector3{X: radius}, VelocityKMPS: Vector3{Y: speed}}
	for i := 0; i < ReplaySampleCount-1; i++ {
		state = twoBodyStep(state, ReplayStep.Seconds())
	}
	angle := speed / radius * ReplayStep.Seconds() * (ReplaySampleCount - 1)
	expected := Vector3{X: radius * math.Cos(angle), Y: radius * math.Sin(angle)}
	errorKM := subtract(state.PositionKM, expected)
	if math.Sqrt(dot(errorKM, errorKM)) > 1e-5 {
		t.Fatalf("integrated position %+v differs from circular solution %+v", state.PositionKM, expected)
	}
}

func TestReplayImpulseOffsetsInitialAndPerturbedState(t *testing.T) {
	initial := CartesianState{PositionKM: Vector3{X: 7000}, VelocityKMPS: Vector3{Y: 7.5}}
	for _, delta := range []float64{-0.5, 0, 0.5} {
		offsets, err := ReplayImpulseOffsets(initial, delta)
		if err != nil {
			t.Fatal(err)
		}
		if offsets[0].PositionKM != (Vector3{}) || math.Abs(offsets[0].VelocityKMPS.Y*1000-delta) > 1e-9 {
			t.Fatalf("wrong impulse at %g m/s: %+v", delta, offsets[0])
		}
		if want := delta * ReplayStep.Seconds() / 1000; math.Abs(offsets[1].PositionKM.Y-want) > 1e-6 {
			t.Fatalf("wrong first interval at %g m/s: %+v", delta, offsets[1])
		}
		if delta == 0 && offsets != ([ReplaySampleCount]CartesianState{}) {
			t.Fatal("zero delta-v must produce zero offsets throughout")
		}
	}
	if _, err := ReplayImpulseOffsets(initial, 0.501); err == nil {
		t.Fatal("expected out-of-range impulse to fail")
	}
}
