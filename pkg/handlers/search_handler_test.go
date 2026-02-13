package handlers

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/common"
	"github.com/pixelvide/kube-sentinel/pkg/handlers/resources"
)

func TestSearchPerformance(t *testing.T) {
	// Backup original search funcs
	originalSearchFuncs := make(map[string]func(c *gin.Context, query string, limit int64) ([]common.SearchResult, error))
	for k, v := range resources.SearchFuncs {
		originalSearchFuncs[k] = v
	}
	defer func() {
		// Restore original search funcs
		resources.SearchFuncs = originalSearchFuncs
	}()

	// Clear existing search funcs and add slow mocks
	resources.SearchFuncs = make(map[string]func(c *gin.Context, query string, limit int64) ([]common.SearchResult, error))

	mockLatency := 100 * time.Millisecond
	numResources := 5

	for i := 0; i < numResources; i++ {
		resources.RegisterSearchFunc("resource"+string(rune('A'+i)), func(c *gin.Context, query string, limit int64) ([]common.SearchResult, error) {
			time.Sleep(mockLatency)
			return []common.SearchResult{{Name: "test"}}, nil
		})
	}

	h := NewSearchHandler()
	c, _ := gin.CreateTestContext(nil)

	start := time.Now()
	_, err := h.Search(c, "test", 10)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	t.Logf("Search took %v", duration)

	// If it takes longer than sum of latencies, it's sequential.
	// If it takes close to max latency, it's parallel.
	// With 5 resources * 100ms, sequential is ~500ms. Parallel should be ~100ms + overhead.
	// We'll assert that it's faster than sequential (allowing some buffer).
	expectedSequential := time.Duration(numResources) * mockLatency
	if duration >= expectedSequential/2 {
		t.Errorf("Search took too long: %v (expected < %v)", duration, expectedSequential/2)
	} else {
		t.Logf("Search performance test passed: %v", duration)
	}
}
