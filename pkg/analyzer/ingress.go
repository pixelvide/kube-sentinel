package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"

	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IngressClassAnalyzer struct{}

func (a *IngressClassAnalyzer) Name() string { return "DeprecatedIngressClass" }

func (a *IngressClassAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	ing, ok := obj.(*netv1.Ingress)
	if !ok {
		return nil, nil
	}

	annotations := ing.Annotations
	if annotations == nil {
		return nil, nil
	}

	if _, exists := annotations["kubernetes.io/ingress.class"]; exists {
		return []Anomaly{
			{
				Severity:    SeverityMedium,
				Title:       "Deprecated Ingress Class Annotation",
				Message:     "This Ingress resource uses the deprecated 'kubernetes.io/ingress.class' annotation.",
				Remediation: "Move the ingress class name to 'spec.ingressClassName' for better compliance with the modern Networking API.",
				RuleID:      "I-001",
			},
		}, nil
	}

	return nil, nil
}

func init() {
	Register(&IngressClassAnalyzer{})
}
