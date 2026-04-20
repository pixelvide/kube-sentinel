package analyzer

import (
	"context"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TopologySpreadAnalyzer struct{}

func (t *TopologySpreadAnalyzer) Name() string { return "TopologySpreadConstraints" }

func (t *TopologySpreadAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	// For this port, we don't have a reliable clusterID passed down easily without changing many signatures.
	// However, we can use the client itself or just context. But wait, `c` IS the client for the cluster.
	// A simple approach for cache key is just "default" since the analyze call is already scoped to a clientSet in handler.
	// Or better, we can skip complex caching for now or use a global cache if we assume one instance per cluster (which isn't true).
	// Let's use a simplified approach: check environment every time but with a short-circuit if recent.
	// Since we don't have clusterID string here (it was in the reference signature but not ours), we will use a workaround.
	// We can trust the handler to pass the right client. We can use a single cache entry for the "current" cluster context if we assume this instance is reused.
	// Actually, `GlobalAnalyzers` are singletons. So they are shared across requests for ALL clusters.
	// This is tricky. Without `clusterID`, we can't safely cache per cluster in a singleton analyzer.
	// For now, let's skip the "isTopologyAware" check or implement it without caching (expensive) or assume it's true (safe default).
	// Let's implement it without caching for safety, but check only 1 node.

	if !t.checkTopologyEnvironment(ctx, c) {
		return nil, nil
	}

	var constraints []corev1.TopologySpreadConstraint
	var kind string

	switch o := obj.(type) {
	case *appsv1.Deployment:
		constraints = o.Spec.Template.Spec.TopologySpreadConstraints
		kind = "Deployment"
	case *appsv1.StatefulSet:
		constraints = o.Spec.Template.Spec.TopologySpreadConstraints
		kind = "StatefulSet"
	default:
		return nil, nil
	}

	if len(constraints) == 0 {
		return []Anomaly{
			{
				Severity:    SeverityMedium,
				Title:       "Missing Topology Spread Constraints",
				Message:     fmt.Sprintf("This %s does not specify any topology spread constraints.", kind),
				Remediation: "Define 'topologySpreadConstraints' in spec.template.spec to ensure high availability across zones or nodes.",
				RuleID:      "TOP-001",
			},
		}, nil
	}

	return nil, nil
}

func (t *TopologySpreadAnalyzer) checkTopologyEnvironment(ctx context.Context, c client.Client) bool {
	// Optimization: check just 1 node
	var nodeList corev1.NodeList
	if err := c.List(ctx, &nodeList, client.Limit(1)); err != nil {
		return true // Default to true on error
	}

	if len(nodeList.Items) == 0 {
		return false
	}

	node := nodeList.Items[0]
	topologyLabels := []string{
		"topology.kubernetes.io/zone",
		"topology.kubernetes.io/region",
		"failure-domain.beta.kubernetes.io/zone",
		"failure-domain.beta.kubernetes.io/region",
	}

	for _, tl := range topologyLabels {
		if _, ok := node.Labels[tl]; ok {
			return true
		}
	}

	return false
}

type AffinityAnalyzer struct{}

func (a *AffinityAnalyzer) Name() string { return "ConflictingAffinity" }

func (a *AffinityAnalyzer) Analyze(ctx context.Context, c client.Client, promClient *prometheus.Client, obj client.Object) ([]Anomaly, error) {
	var affinity *corev1.Affinity

	switch o := obj.(type) {
	case *appsv1.Deployment:
		affinity = o.Spec.Template.Spec.Affinity
	case *appsv1.StatefulSet:
		affinity = o.Spec.Template.Spec.Affinity
	case *appsv1.DaemonSet:
		affinity = o.Spec.Template.Spec.Affinity
	case *corev1.Pod:
		affinity = o.Spec.Affinity
	default:
		return nil, nil
	}

	if affinity == nil || affinity.PodAffinity == nil || affinity.PodAntiAffinity == nil {
		return nil, nil
	}

	// Helper to extract terms
	getTerms := func(podAffinity *corev1.PodAffinity) []corev1.PodAffinityTerm {
		var terms []corev1.PodAffinityTerm
		if podAffinity == nil {
			return terms
		}
		terms = append(terms, podAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
		for _, p := range podAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			terms = append(terms, p.PodAffinityTerm)
		}
		return terms
	}

	getAntiTerms := func(podAntiAffinity *corev1.PodAntiAffinity) []corev1.PodAffinityTerm {
		var terms []corev1.PodAffinityTerm
		if podAntiAffinity == nil {
			return terms
		}
		terms = append(terms, podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
		for _, p := range podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			terms = append(terms, p.PodAffinityTerm)
		}
		return terms
	}

	affinityTerms := getTerms(affinity.PodAffinity)
	antiAffinityTerms := getAntiTerms(affinity.PodAntiAffinity)

	if len(affinityTerms) == 0 || len(antiAffinityTerms) == 0 {
		return nil, nil
	}

	var anomalies []Anomaly

	for _, affTerm := range affinityTerms {
		for _, antiTerm := range antiAffinityTerms {
			if affTerm.TopologyKey == antiTerm.TopologyKey && reflect.DeepEqual(affTerm.LabelSelector, antiTerm.LabelSelector) {
				anomalies = append(anomalies, Anomaly{
					Severity:    SeverityMedium,
					Title:       "Conflicting Affinity Rules",
					Message:     fmt.Sprintf("Pod Affinity and Anti-Affinity rules contradict each other on topology key '%s'.", affTerm.TopologyKey),
					Remediation: "Remove or adjust one of the rules. Requiring a pod to be ON a node (Affinity) and NOT ON that node (Anti-Affinity) with the same criteria is self-canceling.",
					RuleID:      "TOP-002",
				})
			}
		}
	}

	return anomalies, nil
}

func init() {
	Register(&TopologySpreadAnalyzer{})
	Register(&AffinityAnalyzer{})
}
