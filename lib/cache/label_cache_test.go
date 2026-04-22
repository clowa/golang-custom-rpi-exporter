package cache

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestLabelCacheRefreshAndGetLabels(t *testing.T) {
	t.Parallel()

	cache := newLabelCache(func() string { return "test-host" })

	labels := cache.GetLabels()
	if labels.Hostname != "test-host" {
		t.Fatalf("expected hostname test-host, got %q", labels.Hostname)
	}

	if cache.NeedsRefresh() {
		t.Fatal("expected label cache to be refreshed")
	}
}

func TestLabelCacheConcurrentAccess(t *testing.T) {
	t.Parallel()

	var fetchCalls atomic.Int32
	cache := newLabelCache(func() string {
		fetchCalls.Add(1)
		return "concurrent-host"
	})

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				labels := cache.GetLabels()
				if labels.Hostname != "concurrent-host" {
					t.Errorf("unexpected hostname %q", labels.Hostname)
					return
				}
			}
		}()
	}
	wg.Wait()

	if fetchCalls.Load() < 1 {
		t.Fatal("expected hostname function to be called at least once")
	}
}
