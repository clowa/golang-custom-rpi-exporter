package apt_exporter

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestAptCollectorCollectErrorExportsScrapeSuccessZero(t *testing.T) {
	t.Parallel()

	collector := newAptCollector()
	collector.getHostname = func() string { return "test-host" }
	collector.getUpgradable = func() (int, error) { return 0, errors.New("upgradable query failed") }
	collector.getCacheTS = func() (int64, error) { return 0, nil }
	collector.aptCache = newAPTQueryCache(time.Second)

	registry := prometheus.NewRegistry()
	if err := registry.Register(collector); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	expected := strings.NewReader(`# HELP apt_scrape_success Whether the apt collector scrape was successful (1 for success, 0 for failure).
# TYPE apt_scrape_success gauge
apt_scrape_success{instance="test-host"} 0
`)

	if err := testutil.GatherAndCompare(registry, expected, "apt_scrape_success"); err != nil {
		t.Fatalf("gather and compare: %v", err)
	}
}

func TestAPTQueryCacheCachesByTTL(t *testing.T) {
	t.Parallel()

	cache := newAPTQueryCache(100 * time.Millisecond)
	var calls atomic.Int32

	fetch := func() (aptSnapshot, error) {
		return aptSnapshot{
			upgradablePackageCount: int(calls.Add(1)),
			packageCacheTimestamp:  123,
		}, nil
	}

	first, err := cache.get(fetch)
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	second, err := cache.get(fetch)
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}

	if first.upgradablePackageCount != second.upgradablePackageCount {
		t.Fatalf("expected cached value within ttl, got %d and %d", first.upgradablePackageCount, second.upgradablePackageCount)
	}

	time.Sleep(120 * time.Millisecond)
	third, err := cache.get(fetch)
	if err != nil {
		t.Fatalf("third fetch failed: %v", err)
	}

	if third.upgradablePackageCount == second.upgradablePackageCount {
		t.Fatalf("expected refreshed value after ttl, still %d", third.upgradablePackageCount)
	}
}

func TestAPTQueryCacheReturnsStaleValueOnRefreshError(t *testing.T) {
	t.Parallel()

	cache := newAPTQueryCache(50 * time.Millisecond)

	_, err := cache.get(func() (aptSnapshot, error) {
		return aptSnapshot{upgradablePackageCount: 4, packageCacheTimestamp: 100}, nil
	})
	if err != nil {
		t.Fatalf("seed fetch failed: %v", err)
	}

	time.Sleep(60 * time.Millisecond)

	value, err := cache.get(func() (aptSnapshot, error) {
		return aptSnapshot{}, errors.New("refresh failed")
	})
	if err == nil {
		t.Fatal("expected refresh error")
	}

	if value.upgradablePackageCount != 4 || value.packageCacheTimestamp != 100 {
		t.Fatalf("expected stale value, got %+v", value)
	}
}
