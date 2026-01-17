package analyzers

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8testing "k8s.io/client-go/testing"
)

func TestTopologySpreadAnalyzer(t *testing.T) {
	scheme := runtime.NewScheme()

	// Deployment with no topologySpreadConstraints
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "test-deployment",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						// Empty spec
					},
				},
			},
		},
	}

	t.Run("Cluster without topology nodes - No Anomaly", func(t *testing.T) {
		node := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Node",
				"metadata": map[string]interface{}{
					"name": "node1",
					"labels": map[string]interface{}{
						"kubernetes.io/hostname": "node1",
					},
				},
			},
		}

		client := fake.NewSimpleDynamicClient(scheme, node)
		analyzer := &TopologySpreadAnalyzer{}
		anomalies := analyzer.Analyze(deployment, client, "cluster-1")

		if len(anomalies) != 0 {
			t.Errorf("Expected 0 anomalies, got %d", len(anomalies))
		}
	})

	t.Run("Cluster WITH topology nodes - Report Anomaly", func(t *testing.T) {
		node := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Node",
				"metadata": map[string]interface{}{
					"name": "node1",
					"labels": map[string]interface{}{
						"topology.kubernetes.io/zone": "us-east-1a",
					},
				},
			},
		}

		client := fake.NewSimpleDynamicClient(scheme, node)
		analyzer := &TopologySpreadAnalyzer{}
		anomalies := analyzer.Analyze(deployment, client, "cluster-with-topo")

		if len(anomalies) == 0 {
			t.Errorf("Expected anomalies, got 0")
		} else {
			if anomalies[0].Message != "Missing Topology Spread Constraints" {
				t.Errorf("Expected message 'Missing Topology Spread Constraints', got '%s'", anomalies[0].Message)
			}
		}
	})

	t.Run("Caching Logic", func(t *testing.T) {
		// 1. Start with non-topology cluster
		nodeNoTopo := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Node",
				"metadata":   map[string]interface{}{"name": "node1"},
			},
		}
		client := fake.NewSimpleDynamicClient(scheme, nodeNoTopo)
		analyzer := &TopologySpreadAnalyzer{}

		// First check -> False (cached)
		if len(analyzer.Analyze(deployment, client, "cluster-cache-test")) != 0 {
			t.Errorf("Expected 0 anomalies in initial check")
		}

		// 2. Manipulate internals to simulate cache validity
		analyzer.mu.Lock()
		if analyzer.cache == nil {
			analyzer.cache = make(map[string]topologyCacheEntry)
		}
		analyzer.cache["cluster-cache-test"] = topologyCacheEntry{
			isTopologyAware: true,
			lastChecked:     time.Now(),
		}
		analyzer.mu.Unlock()

		// Even though client has a node with NO topology labels, it should return anomaly because cache says "true"
		anomalies := analyzer.Analyze(deployment, client, "cluster-cache-test")
		if len(anomalies) == 0 {
			t.Errorf("Expected anomalies due to cached 'true' state")
		}
	})

	t.Run("Error Listing Nodes - Default to True", func(t *testing.T) {
		gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
		client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
			gvr: "NodeList",
		})

		// Inject reactor to force error on list nodes
		client.PrependReactor("list", "nodes", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, context.DeadlineExceeded
		})

		analyzer := &TopologySpreadAnalyzer{}
		// Analyze should return anomalies because checkTopologyEnvironment returns true on error
		anomalies := analyzer.Analyze(deployment, client, "cluster-error-test")

		if len(anomalies) == 0 {
			t.Errorf("Expected anomalies when listing nodes fails (fail-safe defaulted to true)")
		}
	})

	t.Run("Multi-Context Caching", func(t *testing.T) {
		// Setup two clusters: one with topology, one without
		nodeTopo := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1", "kind": "Node",
				"metadata": map[string]interface{}{"name": "node1", "labels": map[string]interface{}{"topology.kubernetes.io/zone": "us-east-1a"}},
			},
		}
		clientTopo := fake.NewSimpleDynamicClient(scheme, nodeTopo)

		nodeNoTopo := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1", "kind": "Node",
				"metadata": map[string]interface{}{"name": "node2"},
			},
		}
		clientNoTopo := fake.NewSimpleDynamicClient(scheme, nodeNoTopo)

		analyzer := &TopologySpreadAnalyzer{}

		// 1. Analyze Cluster A (Topology)
		anomaliesA := analyzer.Analyze(deployment, clientTopo, "cluster-A")
		if len(anomaliesA) == 0 {
			t.Errorf("Expected anomalies for Cluster A")
		}

		// 2. Analyze Cluster B (No Topology)
		anomaliesB := analyzer.Analyze(deployment, clientNoTopo, "cluster-B")
		if len(anomaliesB) != 0 {
			t.Errorf("Expected 0 anomalies for Cluster B")
		}

		// 3. Verify Cache separation
		analyzer.mu.Lock()
		entryA := analyzer.cache["cluster-A"]
		entryB := analyzer.cache["cluster-B"]
		analyzer.mu.Unlock()

		if !entryA.isTopologyAware {
			t.Errorf("Cache for Cluster A should be topology aware")
		}
		if entryB.isTopologyAware {
			t.Errorf("Cache for Cluster B should NOT be topology aware")
		}
	})
}
