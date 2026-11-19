package orbit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func BenchmarkScreenOMMCatalog(b *testing.B) {
	source := testCatalogSource(time.Date(2000, time.June, 28, 0, 0, 0, 0, time.UTC))
	for _, workload := range []struct {
		objects int
		samples int
	}{{2, 61}, {25, 61}, {100, 61}, {100, 1441}} {
		b.Run(fmt.Sprintf("objects_%d/samples_%d", workload.objects, workload.samples), func(b *testing.B) {
			elements := make([]string, workload.objects)
			for i := range elements {
				record := strings.Replace(vanguardOMM, `"NORAD_CAT_ID": 5`, fmt.Sprintf(`"NORAD_CAT_ID": %d`, 1000+i), 1)
				elements[i] = strings.Replace(record, `"MEAN_ANOMALY": 19.3264`,
					fmt.Sprintf(`"MEAN_ANOMALY": %.8f`, float64(i)*360/float64(workload.objects)), 1)
			}
			records, err := ReadOMMCatalog(strings.NewReader("["+strings.Join(elements, ",")+"]"), source, 7*24*time.Hour)
			if err != nil {
				b.Fatal(err)
			}
			const step = 60
			epoch := records[0].EpochUTC
			config := ScreeningConfig{
				StartUTC: epoch, EndUTC: epoch.Add(time.Duration((workload.samples-1)*step) * time.Second),
				ThresholdMeters: 10_000, CoarseStepSeconds: step,
			}
			pairs := workload.objects * (workload.objects - 1) / 2
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := ScreenOMMCatalog(records, config)
				if err != nil {
					b.Fatal(err)
				}
				if result.ObjectsConsidered != workload.objects || result.PairsEvaluated != pairs || result.SampleCount != workload.samples {
					b.Fatalf("unexpected workload: objects=%d pairs=%d samples=%d", result.ObjectsConsidered, result.PairsEvaluated, result.SampleCount)
				}
			}
			b.ReportMetric(float64(pairs), "pairs/op")
		})
	}
}
