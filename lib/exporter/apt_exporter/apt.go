package apt_exporter

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/arduino/go-apt-client"
	"github.com/clowa/golang-custom-rpi-exporter/lib/cache"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const defaultAPTCacheTTL = 30 * time.Second

type aptSnapshot struct {
	upgradablePackageCount int
	packageCacheTimestamp  int64
}

type aptQueryCache struct {
	mu        sync.RWMutex
	ttl       time.Duration
	snapshot  aptSnapshot
	fetchedAt time.Time
	hasValue  bool
}

func newAPTQueryCache(ttl time.Duration) *aptQueryCache {
	return &aptQueryCache{ttl: ttl}
}

func (c *aptQueryCache) get(fetch func() (aptSnapshot, error)) (aptSnapshot, error) {
	now := time.Now()

	c.mu.RLock()
	if c.hasValue && now.Sub(c.fetchedAt) < c.ttl {
		snapshot := c.snapshot
		c.mu.RUnlock()
		return snapshot, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.hasValue && time.Since(c.fetchedAt) < c.ttl {
		return c.snapshot, nil
	}

	snapshot, err := fetch()
	if err != nil {
		if c.hasValue {
			return c.snapshot, err
		}

		return aptSnapshot{}, err
	}

	c.snapshot = snapshot
	c.fetchedAt = now
	c.hasValue = true

	return snapshot, nil
}

func (c *aptQueryCache) hasSnapshot() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.hasValue
}

type aptCollector struct {
	packageCacheTimestamp  *prometheus.Desc
	upgradablePackageCount *prometheus.Desc
	scrapeSuccess          *prometheus.Desc

	aptCache      *aptQueryCache
	getHostname   func() string
	getUpgradable func() (int, error)
	getCacheTS    func() (int64, error)
}

func Register() {
	collector := newAptCollector()
	prometheus.MustRegister(collector)
}

// You must create a constructor for you collector that
// initializes every descriptor and returns a pointer to the collector
func newAptCollector() *aptCollector {
	const namespace = "apt"

	return &aptCollector{
		packageCacheTimestamp: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "package_cache_timestamp_seconds"),
			"Unix timestamp of the package cache in seconds.",
			[]string{"instance"}, nil,
		),
		upgradablePackageCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "upgradable_packages"),
			"Number of upgradable packages.",
			[]string{"instance"}, nil,
		),
		scrapeSuccess: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "scrape_success"),
			"Whether the apt collector scrape was successful (1 for success, 0 for failure).",
			[]string{"instance"}, nil,
		),
		aptCache:      newAPTQueryCache(defaultAPTCacheTTL),
		getHostname:   func() string { return cache.GetLabelCache().GetLabels().Hostname },
		getUpgradable: getUpgradablePackageCount,
		getCacheTS:    getPackageCacheTimestamps,
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *aptCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.packageCacheTimestamp
	ch <- collector.upgradablePackageCount
	ch <- collector.scrapeSuccess
}

// Collect implements required collect function for all promehteus collectors
func (collector *aptCollector) Collect(ch chan<- prometheus.Metric) {
	instance := collector.getHostname()

	scrapeSuccess := 1.0
	snapshot, err := collector.aptCache.get(func() (aptSnapshot, error) {
		upgradablePackageCount, upgradableErr := collector.getUpgradable()
		if upgradableErr != nil {
			return aptSnapshot{}, upgradableErr
		}

		packageCacheTimestamps, packageCacheErr := collector.getCacheTS()
		if packageCacheErr != nil {
			return aptSnapshot{}, packageCacheErr
		}

		return aptSnapshot{
			upgradablePackageCount: upgradablePackageCount,
			packageCacheTimestamp:  packageCacheTimestamps,
		}, nil
	})
	if err != nil {
		scrapeSuccess = 0
		log.WithError(err).Error("apt collector scrape failed")
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	if err == nil || collector.aptCache.hasSnapshot() {
		ch <- prometheus.MustNewConstMetric(collector.packageCacheTimestamp, prometheus.GaugeValue, float64(snapshot.packageCacheTimestamp), instance)
		ch <- prometheus.MustNewConstMetric(collector.upgradablePackageCount, prometheus.GaugeValue, float64(snapshot.upgradablePackageCount), instance)
	}

	ch <- prometheus.MustNewConstMetric(collector.scrapeSuccess, prometheus.GaugeValue, scrapeSuccess, instance)
}

// getUpgradablePackageCount returns the number of upgradable packages
func getUpgradablePackageCount() (int, error) {
	upgradablePackages, err := apt.ListUpgradable()
	if err != nil {
		return 0, err
	}

	return len(upgradablePackages), nil
}

// getPackageCacheTimestamps returns the timestamp of the package cache
func getPackageCacheTimestamps() (int64, error) {
	const packageCacheTimestampsFilePath = "/var/lib/apt/lists"

	// Get File Stats
	stats, err := os.Stat(packageCacheTimestampsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("package cache path %q does not exist: %w", packageCacheTimestampsFilePath, err)
		}

		return 0, fmt.Errorf("failed to stat package cache path %q: %w", packageCacheTimestampsFilePath, err)
	}

	return stats.ModTime().Unix(), nil
}
