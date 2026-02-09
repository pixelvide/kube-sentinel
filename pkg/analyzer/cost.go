package analyzer

import (
	"context"
	"fmt"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CostOptimizationAnalyzer struct{}

func (a *CostOptimizationAnalyzer) Name() string {
	return "CostOptimization"
}

func (a *CostOptimizationAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	if promClient == nil {
		return nil, nil
	}

	var containers []corev1.Container
	var name string
	var namespace string
	var replicas int32 = 1

	switch o := obj.(type) {
	case *appsv1.Deployment:
		containers = o.Spec.Template.Spec.Containers
		name = o.Name
		namespace = o.Namespace
		if o.Status.Replicas > 0 {
			replicas = o.Status.Replicas
		}
	case *appsv1.StatefulSet:
		containers = o.Spec.Template.Spec.Containers
		name = o.Name
		namespace = o.Namespace
		if o.Status.Replicas > 0 {
			replicas = o.Status.Replicas
		}
	case *appsv1.DaemonSet:
		containers = o.Spec.Template.Spec.Containers
		name = o.Name
		namespace = o.Namespace
		if o.Status.CurrentNumberScheduled > 0 {
			replicas = o.Status.CurrentNumberScheduled
		}
	case *corev1.Pod:
		containers = o.Spec.Containers
		name = o.Name
		namespace = o.Namespace
		replicas = 1
	default:
		return nil, nil
	}

	var anomalies []Anomaly

	// Use 24h history for analysis
	duration := "24h"

	for _, container := range containers {
		// Fetch metrics for this container
		// We use the object name as prefix. This works for Deployments/StatefulSets/DaemonSets/Pods
		metrics, err := promClient.GetPodMetrics(ctx, namespace, name, container.Name, duration)
		if err != nil {
			// If metrics are missing, we just skip analysis for this container
			continue
		}

		// Analyze CPU
		cpuReq := container.Resources.Requests.Cpu()
		cpuLim := container.Resources.Limits.Cpu()

		maxCPUUsageTotal := getMaxUsage(metrics.CPU) // In cores (rate of seconds)
		// Average per pod. Note: GetPodMetrics returns SUM of all replicas, so we divide by replica count to get average usage per pod.
		maxCPUUsage := maxCPUUsageTotal / float64(replicas)

		if !cpuReq.IsZero() {
			reqVal := cpuReq.AsApproximateFloat64()
			// Only flag over-provisioning if usage is really low (e.g. < 50%) AND request is somewhat significant (> 10m) to avoid noise on tiny containers
			if reqVal > 0.01 && maxCPUUsage < reqVal*0.5 {
				anomalies = append(anomalies, Anomaly{
					Severity:    SeverityLow,
					Title:       "CPU Over-provisioned",
					Message:     fmt.Sprintf("Container '%s' requested %s CPU (per pod) but max usage was %.3f cores (avg per pod) (%.0f%% of request).", container.Name, cpuReq.String(), maxCPUUsage, (maxCPUUsage/reqVal)*100),
					Remediation: fmt.Sprintf("Consider reducing CPU request to around %.3f cores to save cost.", maxCPUUsage*1.2), // Buffer 20%
					RuleID:      "COST-001",
				})
			}
		}

		if !cpuLim.IsZero() {
			limVal := cpuLim.AsApproximateFloat64()
			if maxCPUUsage > limVal*0.9 {
				anomalies = append(anomalies, Anomaly{
					Severity:    SeverityMedium,
					Title:       "CPU Limit Risk",
					Message:     fmt.Sprintf("Container '%s' usage reached %.3f cores (avg per pod) (%.0f%% of limit %s).", container.Name, maxCPUUsage, (maxCPUUsage/limVal)*100, cpuLim.String()),
					Remediation: "Consider increasing CPU limit to avoid throttling.",
					RuleID:      "PERF-001",
				})
			}
		}

		// Analyze Memory
		memReq := container.Resources.Requests.Memory()
		memLim := container.Resources.Limits.Memory()

		// GetMemoryUsage in pkg/prometheus/client.go returns value in MiB (bytes / 1024 / 1024)
		maxMemUsageMBTotal := getMaxUsage(metrics.Memory)
		maxMemUsageMB := maxMemUsageMBTotal / float64(replicas)
		// Convert back to bytes to compare with Resource Quantity (which uses bytes)
		maxMemUsageBytes := maxMemUsageMB * 1024 * 1024

		if !memReq.IsZero() {
			reqVal := memReq.AsApproximateFloat64()
			// Check if usage < 50% request
			if reqVal > 10*1024*1024 && maxMemUsageBytes < reqVal*0.5 {
				anomalies = append(anomalies, Anomaly{
					Severity:    SeverityLow,
					Title:       "Memory Over-provisioned",
					Message:     fmt.Sprintf("Container '%s' requested %s Memory (per pod) but max usage was %.0f MiB (avg per pod) (%.0f%% of request).", container.Name, memReq.String(), maxMemUsageMB, (maxMemUsageBytes/reqVal)*100),
					Remediation: fmt.Sprintf("Consider reducing Memory request to around %.0f MiB to save cost.", maxMemUsageMB*1.2),
					RuleID:      "COST-002",
				})
			}
		}

		if !memLim.IsZero() {
			limVal := memLim.AsApproximateFloat64()
			if maxMemUsageBytes > limVal*0.9 {
				anomalies = append(anomalies, Anomaly{
					Severity:    SeverityHigh,
					Title:       "Memory Limit Risk",
					Message:     fmt.Sprintf("Container '%s' usage reached %.0f MiB (avg per pod) (%.0f%% of limit %s).", container.Name, maxMemUsageMB, (maxMemUsageBytes/limVal)*100, memLim.String()),
					Remediation: "Consider increasing Memory limit to avoid OOM kills.",
					RuleID:      "PERF-002",
				})
			}
		}
	}

	return anomalies, nil
}

func getMaxUsage(points []prometheus.UsageDataPoint) float64 {
	maxVal := 0.0
	for _, p := range points {
		if p.Value > maxVal {
			maxVal = p.Value
		}
	}
	return maxVal
}

func init() {
	Register(&CostOptimizationAnalyzer{})
}
