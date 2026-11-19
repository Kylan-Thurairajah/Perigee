package orbit

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

type referencePair struct{ first, second int }

// The reference scans every pair on a fine grid without the production
// candidate selection or refinement. Both paths use the same SGP4 model.
func denseReference(t *testing.T, records []OMMCatalogRecord, start, end time.Time, step time.Duration) map[referencePair]float64 {
	t.Helper()
	propagators := make([]*Propagator, len(records))
	minima := make(map[referencePair]float64)
	for i, record := range records {
		propagator, err := NewPropagatorFromOMM(record.Element)
		if err != nil {
			t.Fatal(err)
		}
		propagators[i] = propagator
	}
	positions := make([]Vector3, len(records))
	for at := start; !at.After(end); at = at.Add(step) {
		for i, propagator := range propagators {
			state, err := propagator.Propagate(at)
			if err != nil {
				t.Fatal(err)
			}
			positions[i] = state.PositionKM
		}
		for i, first := range positions {
			for j := i + 1; j < len(positions); j++ {
				second := positions[j]
				pair := referencePair{records[i].SatelliteNumber, records[j].SatelliteNumber}
				distance := 1000 * math.Hypot(math.Hypot(first.X-second.X, first.Y-second.Y), first.Z-second.Z)
				if previous, ok := minima[pair]; !ok || distance < previous {
					minima[pair] = distance
				}
			}
		}
	}
	return minima
}

func TestScreeningPairRecall(t *testing.T) {
	// Six synthetic crossing pairs derived from the Vallado OMM fixture.
	// Spread each pair around the orbit and vary the second object's anomaly,
	// mean motion, and ascending node to obtain different miss distances.
	var elements []string
	for i := 0; i < 6; i++ {
		for j := 0; j < 2; j++ {
			record := strings.NewReplacer(
				`"NORAD_CAT_ID": 5`, fmt.Sprintf(`"NORAD_CAT_ID": %d`, 2000+i*2+j),
				`"MEAN_ANOMALY": 19.3264`, fmt.Sprintf(`"MEAN_ANOMALY": %.8f`, 19+55*float64(i)-float64(j)*(0.045+0.008*float64(i))),
				`"MEAN_MOTION": 10.82419157`, fmt.Sprintf(`"MEAN_MOTION": %.8f`, 10.82419157+0.01*float64(j)),
				`"RA_OF_ASC_NODE": 348.7242`, fmt.Sprintf(`"RA_OF_ASC_NODE": %.8f`, 348.7242+0.012*float64(i*j)),
			).Replace(vanguardOMM)
			elements = append(elements, record)
		}
	}
	source := testCatalogSource(time.Date(2000, time.June, 28, 0, 0, 0, 0, time.UTC))
	records, err := ReadOMMCatalog(strings.NewReader("["+strings.Join(elements, ",")+"]"), source, 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	start := records[0].EpochUTC
	end := start.Add(time.Hour)
	reference := denseReference(t, records, start, end, 500*time.Millisecond)
	check := denseReference(t, records, start, end, time.Second)
	for i := 0; i < 6; i++ {
		pair := referencePair{2000 + i*2, 2001 + i*2}
		t.Logf("reference pair %d/%d: %.1f m", pair.first, pair.second, reference[pair])
	}
	for _, fixture := range []struct {
		threshold float64
		pairs     int
	}{{5000, 2}, {10000, 6}, {20000, 6}} {
		threshold := fixture.threshold
		want := make(map[referencePair]bool)
		for pair, distance := range reference {
			if (distance <= threshold) != (check[pair] <= threshold) {
				t.Fatalf("reference grids disagree for %+v at %.0f m", pair, threshold)
			}
			if distance <= threshold {
				want[pair] = true
			}
		}
		if len(want) != fixture.pairs || len(reference) != 66 {
			t.Fatalf("threshold %.0f m: reference found %d/%d pairs, expected %d/66", threshold, len(want), len(reference), fixture.pairs)
		}
		for _, step := range []int{5, 30, 60, 300, 600, 1800, 3600} {
			result, err := ScreenOMMCatalog(records, ScreeningConfig{
				StartUTC: start, EndUTC: end, ThresholdMeters: threshold, CoarseStepSeconds: step,
			})
			if err != nil {
				t.Fatal(err)
			}
			hits, extras := 0, 0
			seen := make(map[referencePair]bool)
			for _, item := range result.Encounters {
				pair := referencePair{item.ObjectANORADID, item.ObjectBNORADID}
				if _, ok := reference[pair]; !ok || seen[pair] {
					t.Fatalf("unexpected or duplicate result pair: %+v", pair)
				}
				seen[pair] = true
				if want[pair] {
					hits++
				} else {
					extras++
				}
			}
			t.Logf("threshold=%5.0f m step=%4d s reference=%d recovered=%d extras=%d", threshold, step, len(want), hits, extras)
			if (step == 5 || step == 60) && (hits != len(want) || extras != 0) {
				t.Errorf("%d-second fixture regression at %.0f m: recovered %d/%d, additional pairs %d", step, threshold, hits, len(want), extras)
			}
		}
	}
}
