package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DefaultNamespaceAnalyzer struct{}

func (a *DefaultNamespaceAnalyzer) Name() string {
	return "DefaultNamespace"
}

func (a *DefaultNamespaceAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	if obj.GetNamespace() == "default" {
		return []Anomaly{
			{
				Severity:    SeverityLow,
				Title:       "Default Namespace Usage",
				Message:     fmt.Sprintf("Resource %s is in the default namespace", obj.GetName()),
				Remediation: "Consider moving important workloads to specific namespaces for better isolation and management.",
				RuleID:      "G-001",
			},
		}, nil
	}
	return nil, nil
}

func init() {
	Register(&DefaultNamespaceAnalyzer{})
}
