package orbit

import (
	"fmt"
	"math"
	"time"
)

const (
	ReplayStep        = 10 * time.Second
	ReplaySampleCount = 61
	MaxReplayDeltaV   = 0.5 // m/s
	earthMuKM3PerS2   = 398600.4418
)

type CartesianState struct {
	PositionKM   Vector3
	VelocityKMPS Vector3
}

// ReplayImpulseOffsets integrates the same initial state twice, with and without
// an initial impulse parallel to velocity, to estimate the incremental change.
func ReplayImpulseOffsets(initial CartesianState, deltaVMPS float64) ([ReplaySampleCount]CartesianState, error) {
	var offsets [ReplaySampleCount]CartesianState
	speed := math.Sqrt(dot(initial.VelocityKMPS, initial.VelocityKMPS))
	if !validTwoBodyState(initial) || speed <= 0 || !finiteScalar(deltaVMPS) || math.Abs(deltaVMPS) > MaxReplayDeltaV {
		return offsets, fmt.Errorf("invalid orbital state or impulse")
	}
	baseline, perturbed := initial, initial
	perturbed.VelocityKMPS = add(initial.VelocityKMPS, scale(initial.VelocityKMPS, deltaVMPS/(1000*speed)))
	for i := range offsets {
		if i > 0 {
			baseline = twoBodyStep(baseline, ReplayStep.Seconds())
			perturbed = twoBodyStep(perturbed, ReplayStep.Seconds())
		}
		if !validTwoBodyState(baseline) || !validTwoBodyState(perturbed) {
			return offsets, fmt.Errorf("two-body integration left the supported state range")
		}
		offsets[i] = CartesianState{
			PositionKM:   subtract(perturbed.PositionKM, baseline.PositionKM),
			VelocityKMPS: subtract(perturbed.VelocityKMPS, baseline.VelocityKMPS),
		}
	}
	return offsets, nil
}

func validTwoBodyState(state CartesianState) bool {
	radiusSquared := dot(state.PositionKM, state.PositionKM)
	speedSquared := dot(state.VelocityKMPS, state.VelocityKMPS)
	return radiusSquared > 0 && finiteScalar(radiusSquared) && finiteScalar(speedSquared)
}

func finiteScalar(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func twoBodyStep(state CartesianState, seconds float64) CartesianState {
	derivative := func(state CartesianState) CartesianState {
		radius := math.Sqrt(dot(state.PositionKM, state.PositionKM))
		return CartesianState{state.VelocityKMPS, scale(state.PositionKM, -earthMuKM3PerS2/(radius*radius*radius))}
	}
	advance := func(state, rate CartesianState, step float64) CartesianState {
		return CartesianState{add(state.PositionKM, scale(rate.PositionKM, step)), add(state.VelocityKMPS, scale(rate.VelocityKMPS, step))}
	}
	// Classical fourth-order Runge–Kutta integration.
	k1 := derivative(state)
	k2 := derivative(advance(state, k1, seconds/2))
	k3 := derivative(advance(state, k2, seconds/2))
	k4 := derivative(advance(state, k3, seconds))
	weighted := func(a, b, c, d Vector3) Vector3 {
		return add(add(a, d), scale(add(b, c), 2))
	}
	return advance(state, CartesianState{
		weighted(k1.PositionKM, k2.PositionKM, k3.PositionKM, k4.PositionKM),
		weighted(k1.VelocityKMPS, k2.VelocityKMPS, k3.VelocityKMPS, k4.VelocityKMPS),
	}, seconds/6)
}
