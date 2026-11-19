package orbit

import (
	"fmt"
	"strings"
	"time"
)

type CatalogSource struct {
	Name           string    `json:"name"`
	URL            string    `json:"url,omitempty"`
	SnapshotID     string    `json:"snapshotId"`
	RetrievedAtUTC time.Time `json:"retrievedAtUtc"`
}

func validateCatalogSource(source CatalogSource, maxEpochAge time.Duration, format string) (CatalogSource, error) {
	if strings.TrimSpace(source.Name) == "" {
		return CatalogSource{}, fmt.Errorf("catalog source name is required")
	}
	if strings.TrimSpace(source.SnapshotID) == "" {
		return CatalogSource{}, fmt.Errorf("catalog snapshot ID is required")
	}
	if source.RetrievedAtUTC.IsZero() {
		return CatalogSource{}, fmt.Errorf("catalog retrieval time is required")
	}
	if maxEpochAge <= 0 {
		return CatalogSource{}, fmt.Errorf("maximum %s epoch age must be positive", format)
	}
	source.RetrievedAtUTC = source.RetrievedAtUTC.UTC()
	return source, nil
}
