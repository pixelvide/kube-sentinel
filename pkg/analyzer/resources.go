package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceLimitsAnalyzer struct{}

func (a *ResourceLimitsAnalyzer) Name() string {
	return "ResourceLimits"
}

func (a *ResourceLimitsAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var containers []corev1.Container
	var name string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		containers = o.Spec.Template.Spec.Containers
		name = o.Name
	case *appsv1.StatefulSet:
		containers = o.Spec.Template.Spec.Containers
		name = o.Name
	case *appsv1.DaemonSet:
		containers = o.Spec.Template.Spec.Containers
		name = o.Name
	case *corev1.Pod:
		containers = o.Spec.Containers
		name = o.Name
	default:
		return nil, nil
	}

	var anomalies []Anomaly

	for _, c := range containers {
		if c.Resources.Limits == nil {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Missing Resource Limits",
				Message:     fmt.Sprintf("Container '%s' in '%s' has no resource limits defined.", c.Name, name),
				Remediation: "Set CPU and Memory limits to prevent resource contention and ensure stable scheduling.",
				RuleID:      "R-001",
			})
			continue
		}

		if _, ok := c.Resources.Limits[corev1.ResourceCPU]; !ok {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Missing CPU Limit",
				Message:     fmt.Sprintf("Container '%s' in '%s' has no CPU limit defined.", c.Name, name),
				Remediation: "Set a CPU limit to prevent this container from starving others.",
				RuleID:      "R-002",
			})
		}
		if _, ok := c.Resources.Limits[corev1.ResourceMemory]; !ok {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Missing Memory Limit",
				Message:     fmt.Sprintf("Container '%s' in '%s' has no Memory limit defined.", c.Name, name),
				Remediation: "Set a Memory limit to prevent OutOfMemory (OOM) kills affecting the node.",
				RuleID:      "R-003",
			})
		}
	}

	return anomalies, nil
}

func init() {
	Register(&ResourceLimitsAnalyzer{})
}
