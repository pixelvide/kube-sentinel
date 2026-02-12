package analyzer

import (
	"context"
	"sync"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	analyzers []Analyzer
	mu        sync.RWMutex
)

func Register(a Analyzer) {
	mu.Lock()
	defer mu.Unlock()
	analyzers = append(analyzers, a)
}

func Analyze(ctx context.Context, k8sClient client.Client, promClient *prometheus.Client, obj client.Object) *ResourceAnalysis {
	mu.RLock()
	defer mu.RUnlock()

	var anomalies []Anomaly
	for _, a := range analyzers {
		results, err := a.Analyze(ctx, k8sClient, promClient, obj)
		if err != nil {
			klog.Errorf("Analyzer %s failed: %v", a.Name(), err)
			continue
		}
		anomalies = append(anomalies, results...)
	}

	analysis := &ResourceAnalysis{
		Anomalies: anomalies,
		Score:     100, // Placeholder
	}

	if len(anomalies) > 0 {
		analysis.Summary = "Anomalies detected"
		// Simple score reduction logic
		for _, anomaly := range anomalies {
			switch anomaly.Severity {
			case SeverityCritical:
				analysis.Score -= 20
			case SeverityHigh:
				analysis.Score -= 10
			case SeverityMedium:
				analysis.Score -= 5
			case SeverityLow:
				analysis.Score -= 2
			}
		}
		if analysis.Score < 0 {
			analysis.Score = 0
		}
	} else {
		analysis.Summary = "No anomalies detected"
	}

	return analysis
}
