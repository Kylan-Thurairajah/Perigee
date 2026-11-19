package orbit

import (
	"fmt"
	"time"

	"github.com/akhenakh/sgp4"
)

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

type TEMEState struct {
	AtUTC        time.Time
	PositionKM   Vector3
	VelocityKMPS Vector3
}

type Propagator struct {
	tle *sgp4.TLE
}

func NewPropagator(line1, line2 string) (*Propagator, error) {
	tle, err := sgp4.ParseTLELines([]string{line1, line2})
	if err != nil {
		return nil, fmt.Errorf("parse TLE: %w", err)
	}
	return &Propagator{tle: tle}, nil
}

func NewPropagatorFromOMM(omm sgp4.OMM) (*Propagator, error) {
	if reason := ommPropagationLimitation(omm); reason != "" {
		return nil, fmt.Errorf("OMM is not supported for SGP4 propagation: %s", reason)
	}
	tle, err := omm.ToTLE()
	if err != nil {
		return nil, fmt.Errorf("convert OMM elements for SGP4: %w", err)
	}
	if _, err := tle.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize SGP4 from OMM: %w", err)
	}
	return &Propagator{tle: tle}, nil
}

func (p *Propagator) Propagate(at time.Time) (TEMEState, error) {
	if p == nil || p.tle == nil {
		return TEMEState{}, fmt.Errorf("SGP4 propagator is not initialized")
	}

	at = at.UTC()
	eci, err := p.tle.FindPositionAtTime(at)
	if err != nil {
		return TEMEState{}, fmt.Errorf("SGP4 propagation failed: %w", err)
	}
	return TEMEState{
		AtUTC: at,
		PositionKM: Vector3{
			X: eci.Position.X,
			Y: eci.Position.Y,
			Z: eci.Position.Z,
		},
		VelocityKMPS: Vector3{
			X: eci.Velocity.X,
			Y: eci.Velocity.Y,
			Z: eci.Velocity.Z,
		},
	}, nil
}
