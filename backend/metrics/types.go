package metrics

import (
	"time"
)

// UsageDataPoint represents a single time point in usage metrics
type UsageDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// PodMetrics contains metrics for a specific pod
type PodMetrics struct {
	CPU      []UsageDataPoint `json:"cpu"`
	Memory   []UsageDataPoint `json:"memory"`
	Fallback bool             `json:"fallback"`
}
