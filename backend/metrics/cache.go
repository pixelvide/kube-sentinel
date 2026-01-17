package metrics

import (
	"sync"
	"time"
)

type MetricsCache struct {
	cache map[string][]UsageDataPoint
	lock  sync.Mutex
}

func NewMetricsCache() *MetricsCache {
	c := &MetricsCache{
		cache: make(map[string][]UsageDataPoint),
	}
	go c.cleanup()
	return c
}

func (c *MetricsCache) cleanup() {
	for {
		time.Sleep(time.Minute)
		c.lock.Lock()
		cutoff := time.Now().Add(-30 * time.Minute)
		for key, points := range c.cache {
			var filtered []UsageDataPoint
			for _, pt := range points {
				if pt.Timestamp.After(cutoff) {
					filtered = append(filtered, pt)
				}
			}
			if len(filtered) > 0 {
				c.cache[key] = filtered
			} else {
				delete(c.cache, key)
			}
		}
		c.lock.Unlock()
	}
}

func (c *MetricsCache) Add(key string, value float64, ts time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()

	points := c.cache[key]
	// If the last point is very recent (less than 15s), update it instead of appending
	if len(points) > 0 {
		last := points[len(points)-1]
		if ts.Sub(last.Timestamp) < 15*time.Second {
			points[len(points)-1].Value = value
			c.cache[key] = points
			return
		}
	}

	c.cache[key] = append(points, UsageDataPoint{Timestamp: ts, Value: value})
}

func (c *MetricsCache) Get(key string) []UsageDataPoint {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.cache[key]
}
