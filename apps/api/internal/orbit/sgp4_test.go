package orbit

import (
	"math"
	"testing"
	"time"
)

func TestPropagatorMatchesValladoNearEarthReference(t *testing.T) {
	propagator, err := NewPropagator(
		"1 00005U 58002B   00179.78495062  .00000023  00000-0  28098-4 0  4753",
		"2 00005  34.2682 348.7242 1859667 331.7664  19.3264 10.82419157413667",
	)
	if err != nil {
		t.Fatalf("NewPropagator: %v", err)
	}

	epochUTC := time.Date(2000, time.June, 27, 18, 50, 19, 733568000, time.UTC)
	epoch := epochUTC.In(time.FixedZone("test", -7*60*60))
	tests := []struct {
		name         string
		at           time.Time
		positionKM   Vector3
		velocityKMPS Vector3
	}{
		{
			name:         "epoch",
			at:           epoch,
			positionKM:   Vector3{X: 7022.46529266, Y: -1400.08296755, Z: 0.03995155},
			velocityKMPS: Vector3{X: 1.893841015, Y: 6.405893759, Z: 4.534807250},
		},
		{
			name:         "six hours after epoch",
			at:           epoch.Add(360 * time.Minute),
			positionKM:   Vector3{X: -7154.03120202, Y: -3783.17682504, Z: -3536.19412294},
			velocityKMPS: Vector3{X: 4.741887409, Y: -4.151817765, Z: -2.093935425},
		},
	}

	const positionToleranceKM = 1e-4
	const velocityToleranceKMPS = 1e-6
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := propagator.Propagate(test.at)
			if err != nil {
				t.Fatalf("Propagate: %v", err)
			}
			if !got.AtUTC.Equal(test.at.UTC()) {
				t.Errorf("AtUTC = %s, want %s", got.AtUTC, test.at.UTC())
			}
			if distance(got.PositionKM, test.positionKM) > positionToleranceKM {
				t.Errorf("position = %+v, want %+v (tolerance %g km)", got.PositionKM, test.positionKM, positionToleranceKM)
			}
			if distance(got.VelocityKMPS, test.velocityKMPS) > velocityToleranceKMPS {
				t.Errorf("velocity = %+v, want %+v (tolerance %g km/s)", got.VelocityKMPS, test.velocityKMPS, velocityToleranceKMPS)
			}
		})
	}
}

func TestNewPropagatorRejectsInvalidTLE(t *testing.T) {
	if _, err := NewPropagator("not a TLE", "also not a TLE"); err == nil {
		t.Fatal("NewPropagator accepted invalid TLE lines")
	}
}

func distance(a, b Vector3) float64 {
	return math.Sqrt(
		math.Pow(a.X-b.X, 2) +
			math.Pow(a.Y-b.Y, 2) +
			math.Pow(a.Z-b.Z, 2),
	)
}
