package orbit

import (
	"strings"
	"testing"
	"time"
)

func TestScreenOMMCatalogFindsAndRefinesCloseApproach(t *testing.T) {
	second := strings.Replace(vanguardOMM, `"OBJECT_NAME": "VANGUARD 1"`, `"OBJECT_NAME": "VANGUARD 2"`, 1)
	second = strings.Replace(second, `"NORAD_CAT_ID": 5`, `"NORAD_CAT_ID": 6`, 1)
	second = strings.Replace(second, `"MEAN_MOTION": 10.82419157`, `"MEAN_MOTION": 10.83419157`, 1)
	second = strings.Replace(second, `"MEAN_ANOMALY": 19.3264`, `"MEAN_ANOMALY": 19.2764`, 1)
	source := testCatalogSource(time.Date(2000, time.June, 28, 0, 0, 0, 0, time.UTC))
	records, err := ReadOMMCatalog(strings.NewReader("["+vanguardOMM+","+second+"]"), source, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog: %v", err)
	}

	epoch := records[0].EpochUTC
	result, err := ScreenOMMCatalog(records, ScreeningConfig{
		StartUTC:          epoch.Add(17 * time.Minute),
		EndUTC:            epoch.Add(23 * time.Minute),
		ThresholdMeters:   10_000,
		CoarseStepSeconds: 60,
	})
	if err != nil {
		t.Fatalf("ScreenOMMCatalog: %v", err)
	}
	if result.SampleCount != 7 || result.PairsEvaluated != 1 || len(result.Encounters) != 1 {
		t.Fatalf("unexpected screen result: %+v", result)
	}
	encounter := result.Encounters[0]
	if encounter.ObjectANORADID != 5 || encounter.ObjectBNORADID != 6 ||
		encounter.MinimumSeparationMeters <= 0 || encounter.MinimumSeparationMeters > 10_000 ||
		encounter.RelativeSpeedMetersPerS <= 0 {
		t.Fatalf("unexpected conjunction result: %+v", encounter)
	}
	if delta := encounter.TCAUTC.Sub(epoch.Add(20 * time.Minute)); delta < -2*time.Minute || delta > 2*time.Minute {
		t.Fatalf("TCA = %s, expected near the constructed phase crossing", encounter.TCAUTC)
	}
}

func TestScreenOMMCatalogRejectsStaleUnsupportedAndUnboundedRuns(t *testing.T) {
	source := testCatalogSource(time.Date(2000, time.July, 10, 0, 0, 0, 0, time.UTC))
	staleSecond := strings.Replace(vanguardOMM, `"NORAD_CAT_ID": 5`, `"NORAD_CAT_ID": 6`, 1)
	stale, err := ReadOMMCatalog(strings.NewReader("["+vanguardOMM+","+staleSecond+"]"), source, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog stale records: %v", err)
	}
	config := ScreeningConfig{
		StartUTC:          time.Date(2000, time.June, 27, 18, 50, 0, 0, time.UTC),
		EndUTC:            time.Date(2000, time.June, 27, 18, 51, 0, 0, time.UTC),
		ThresholdMeters:   1000,
		CoarseStepSeconds: 10,
	}
	if _, err := ScreenOMMCatalog(stale, config); err == nil {
		t.Fatal("ScreenOMMCatalog accepted stale elements")
	}

	freshSource := testCatalogSource(time.Date(2000, time.June, 28, 0, 0, 0, 0, time.UTC))
	unsupportedJSON := strings.Replace(vanguardOMM, `"REF_FRAME": "TEME"`, `"REF_FRAME": "ITRF"`, 1)
	unsupportedJSON = "[" + unsupportedJSON + "," + strings.Replace(unsupportedJSON, `"NORAD_CAT_ID": 5`, `"NORAD_CAT_ID": 6`, 1) + "]"
	unsupported, err := ReadOMMCatalog(strings.NewReader(unsupportedJSON), freshSource, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog unsupported records: %v", err)
	}
	if _, err := ScreenOMMCatalog(unsupported, config); err == nil {
		t.Fatal("ScreenOMMCatalog accepted unsupported elements")
	}

	valid, err := ReadOMMCatalog(strings.NewReader("["+vanguardOMM+","+strings.Replace(vanguardOMM, `"NORAD_CAT_ID": 5`, `"NORAD_CAT_ID": 6`, 1)+"]"), freshSource, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ReadOMMCatalog valid records: %v", err)
	}
	config.EndUTC = config.StartUTC.Add(24 * time.Hour)
	config.CoarseStepSeconds = 1
	if _, err := ScreenOMMCatalog(valid, config); err == nil {
		t.Fatal("ScreenOMMCatalog accepted an unbounded sample count")
	}
}
