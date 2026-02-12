package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SingleReplicaAnalyzer struct{}

func (a *SingleReplicaAnalyzer) Name() string { return "SingleReplica" }

func (a *SingleReplicaAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var replicas int32
	var kind string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		if o.Spec.Replicas != nil {
			replicas = *o.Spec.Replicas
		} else {
			replicas = 1
		}
		kind = "Deployment"
	case *appsv1.StatefulSet:
		if o.Spec.Replicas != nil {
			replicas = *o.Spec.Replicas
		} else {
			replicas = 1
		}
		kind = "StatefulSet"
	default:
		return nil, nil
	}

	if replicas == 1 {
		return []Anomaly{
			{
				Severity:    SeverityMedium,
				Title:       "Single Replica Detected",
				Message:     fmt.Sprintf("This %s has only 1 replica running.", kind),
				Remediation: "Increase replica count to at least 2 to ensure high availability and zero-downtime rolling updates.",
				RuleID:      "REL-001",
			},
		}, nil
	}

	return nil, nil
}

type ProbeAnalyzer struct{}

func (a *ProbeAnalyzer) Name() string { return "MissingProbes" }

func (a *ProbeAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
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
		if c.LivenessProbe == nil {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Missing Liveness Probe",
				Message:     fmt.Sprintf("Container '%s' in '%s' is missing a liveness probe.", c.Name, name),
				Remediation: "Define a liveness probe to allow Kubernetes to restart the container if it deadlocks or crashes locally.",
				RuleID:      "REL-002",
			})
		}
		if c.ReadinessProbe == nil {
			anomalies = append(anomalies, Anomaly{
				Severity:    SeverityMedium,
				Title:       "Missing Readiness Probe",
				Message:     fmt.Sprintf("Container '%s' in '%s' is missing a readiness probe.", c.Name, name),
				Remediation: "Define a readiness probe to prevent traffic from being sent to the container before it is ready to handle requests.",
				RuleID:      "REL-003",
			})
		}
	}

	return anomalies, nil
}

type MissingPDBAnalyzer struct{}

func (a *MissingPDBAnalyzer) Name() string { return "MissingPDB" }

func (a *MissingPDBAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var selector *metav1.LabelSelector
	var namespace string
	var kind string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		selector = o.Spec.Selector
		namespace = o.Namespace
		kind = "Deployment"
	case *appsv1.StatefulSet:
		selector = o.Spec.Selector
		namespace = o.Namespace
		kind = "StatefulSet"
	default:
		return nil, nil
	}

	if selector == nil || len(selector.MatchLabels) == 0 {
		return nil, nil // Cannot match if no selector
	}

	workloadLabels := selector.MatchLabels

	// List PDBs in the namespace
	var pdbList policyv1.PodDisruptionBudgetList
	if err := c.List(ctx, &pdbList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	hasPDB := false
	for _, pdb := range pdbList.Items {
		if pdb.Spec.Selector == nil {
			continue
		}

		// Check if PDB selector matches Workload labels
		// Simple equality check for now (as per original logic)
		// A robust check would need to verify if PDB selector covers the workload's pods
		// But here we check if PDB selector is a subset or equal to workload selector
		// The original logic checked if PDB selector matches workload matchLabels

		pdbMatch := pdb.Spec.Selector.MatchLabels
		if len(pdbMatch) == 0 {
			continue
		}

		match := true
		for k, v := range pdbMatch {
			if val, ok := workloadLabels[k]; !ok || val != v {
				match = false
				break
			}
		}

		if match {
			hasPDB = true
			break
		}
	}

	if !hasPDB {
		return []Anomaly{
			{
				Severity:    SeverityMedium,
				Title:       "Missing PodDisruptionBudget",
				Message:     fmt.Sprintf("This %s is not covered by any PodDisruptionBudget (PDB).", kind),
				Remediation: "Create a PDB to ensure service availability during voluntary disruptions (e.g. node drains, upgrades).",
				RuleID:      "REL-004",
			},
		}, nil
	}

	return nil, nil
}

func init() {
	Register(&SingleReplicaAnalyzer{})
	Register(&ProbeAnalyzer{})
	Register(&MissingPDBAnalyzer{})
}
