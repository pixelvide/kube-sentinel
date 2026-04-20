package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DanglingServiceAnalyzer struct{}

func (a *DanglingServiceAnalyzer) Name() string { return "DanglingService" }

func (a *DanglingServiceAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		return nil, nil
	}

	if len(svc.Spec.Selector) == 0 {
		return nil, nil
	}

	// List Pods matching the selector
	var podList corev1.PodList
	if err := c.List(ctx, &podList, client.InNamespace(svc.Namespace), client.MatchingLabels(svc.Spec.Selector)); err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return []Anomaly{
			{
				Severity:    SeverityMedium,
				Title:       "Dangling Service Detected",
				Message:     "This Service's selector matches 0 active Pods.",
				Remediation: "Check if the selector labels match the labels on your target Pods/Deployments.",
				RuleID:      "H-001",
			},
		}, nil
	}

	return nil, nil
}

type EmptyNamespaceAnalyzer struct{}

func (a *EmptyNamespaceAnalyzer) Name() string { return "EmptyNamespace" }

func (a *EmptyNamespaceAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil, nil
	}

	name := ns.Name
	// Skip system namespaces
	if name == "kube-system" || name == "kube-public" || name == "kube-node-lease" || name == "default" {
		return nil, nil
	}

	// We need to check exact count.
	// To minimize boilerplate, let's just do explicit calls for common workloads.

	var depList appsv1.DeploymentList
	if err := c.List(ctx, &depList, client.InNamespace(name), client.Limit(1)); err == nil && len(depList.Items) > 0 {
		return nil, nil
	}
	var ssList appsv1.StatefulSetList
	if err := c.List(ctx, &ssList, client.InNamespace(name), client.Limit(1)); err == nil && len(ssList.Items) > 0 {
		return nil, nil
	}
	var dsList appsv1.DaemonSetList
	if err := c.List(ctx, &dsList, client.InNamespace(name), client.Limit(1)); err == nil && len(dsList.Items) > 0 {
		return nil, nil
	}
	var podList corev1.PodList
	if err := c.List(ctx, &podList, client.InNamespace(name), client.Limit(1)); err == nil && len(podList.Items) > 0 {
		return nil, nil
	}
	var svcList corev1.ServiceList
	if err := c.List(ctx, &svcList, client.InNamespace(name), client.Limit(1)); err == nil && len(svcList.Items) > 0 {
		return nil, nil
	}
	var cmList corev1.ConfigMapList
	if err := c.List(ctx, &cmList, client.InNamespace(name)); err == nil {
		for _, cm := range cmList.Items {
			if cm.Name != "kube-root-ca.crt" {
				return nil, nil
			}
		}
	}

	return []Anomaly{
		{
			Severity:    SeverityLow,
			Title:       "Empty Namespace Detected",
			Message:     fmt.Sprintf("Namespace '%s' appears to have no active workloads or services.", name),
			Remediation: "Delete this namespace if it is no longer in use to reduce cluster clutter.",
			RuleID:      "H-002",
		},
	}, nil
}

func init() {
	Register(&DanglingServiceAnalyzer{})
	Register(&EmptyNamespaceAnalyzer{})
}
