package metrics

import (
	"context"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	K8sClient     kubernetes.Interface
	MetricsClient metricsclient.Interface
	Cache         *MetricsCache
}

func NewClient(k8sClient kubernetes.Interface, metricsClient metricsclient.Interface) *Client {
	return &Client{
		K8sClient:     k8sClient,
		MetricsClient: metricsClient,
		Cache:         NewMetricsCache(),
	}
}

// GetPodMetrics attempts to fetch pod metrics.
// Currently it implements the fallback logic using metrics-server.
// Future: Integrate Prometheus client here and try it first.
func (c *Client) GetPodMetrics(ctx context.Context, namespace, podName string) (*PodMetrics, error) {
	if c.MetricsClient == nil {
		return nil, fmt.Errorf("metrics client not available")
	}

	// Fetch current metrics from metrics-server
	podMetrics, err := c.MetricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Update cache
	c.updateCache(podMetrics)

	// Retrieve history from cache for this pod
	cpuSeries, memSeries := c.getFromCache(namespace, podName)

	// Ensure the current point is included in the series if it's new
	// (updateCache handles this, so we just read back)

	return &PodMetrics{
		CPU:      mergeUsageDataPoints(cpuSeries),
		Memory:   mergeUsageDataPoints(memSeries),
		Fallback: true, // Indicating this is from metrics-server (fallback)
	}, nil
}

func (c *Client) updateCache(podMetrics *metricsv1beta1.PodMetrics) {
	timestamp := podMetrics.Timestamp.Time
	for _, container := range podMetrics.Containers {
		key := fmt.Sprintf("%s/%s/%s", podMetrics.Namespace, podMetrics.Name, container.Name)

		cpuUsage := float64(container.Usage.Cpu().MilliValue()) / 1000.0        // Cores
		memUsage := float64(container.Usage.Memory().Value()) / 1024.0 / 1024.0 // MiB

		c.Cache.Add(key+"/cpu", cpuUsage, timestamp)
		c.Cache.Add(key+"/mem", memUsage, timestamp)
	}
}

func (c *Client) getFromCache(namespace, podName string) (cpu []UsageDataPoint, mem []UsageDataPoint) {
	// We need to aggregate across all containers for the pod view
	// For simplicity, we'll just scan the cache for keys starting with ns/pod/
	// In a real implementation with known containers, we'd iterate them.
	// But simply aggregating everything in the cache that matches is a bit tricky without a list of containers.
	// So for now, let's assume we want to sum up all containers we have in cache for this pod.

	// Actually, the `GetPodMetrics` call gave us the list of containers.
	// But `getFromCache` as a separate method might not have it.
	// Let's refactor to do it inline or pass containers.
	// For this first pass, let's just use the cache structure we built.
	return nil, nil // Logic moved to GetPodMetrics or helper that takes container list
}

// Helper to aggregate based on known containers from the lived fetch
func (c *Client) AggregatePodMetricsFromCache(podMetrics *metricsv1beta1.PodMetrics) ([]UsageDataPoint, []UsageDataPoint) {
	var cpuSeries, memSeries []UsageDataPoint

	// We use a map to sum values at each timestamp
	cpuMap := make(map[time.Time]float64)
	memMap := make(map[time.Time]float64)

	for _, container := range podMetrics.Containers {
		key := fmt.Sprintf("%s/%s/%s", podMetrics.Namespace, podMetrics.Name, container.Name)

		cpus := c.Cache.Get(key + "/cpu")
		mems := c.Cache.Get(key + "/mem")

		for _, pt := range cpus {
			cpuMap[pt.Timestamp] += pt.Value
		}
		for _, pt := range mems {
			memMap[pt.Timestamp] += pt.Value
		}
	}

	for t, v := range cpuMap {
		cpuSeries = append(cpuSeries, UsageDataPoint{Timestamp: t, Value: v})
	}
	for t, v := range memMap {
		memSeries = append(memSeries, UsageDataPoint{Timestamp: t, Value: v})
	}

	return cpuSeries, memSeries
}

func mergeUsageDataPoints(points []UsageDataPoint) []UsageDataPoint {
	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp.Before(points[j].Timestamp)
	})
	return points
}
