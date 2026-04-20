package analyzer

import (
	"context"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, client client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error)
}
