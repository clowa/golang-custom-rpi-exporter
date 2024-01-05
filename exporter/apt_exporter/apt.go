package apt_exporter

import (
	"fmt"
	"os"

	"github.com/arduino/go-apt-client"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type aptCollector struct {
	packageCacheTimestamp  *prometheus.Desc
	upgradablePackageCount *prometheus.Desc
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
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *aptCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.packageCacheTimestamp
	ch <- collector.upgradablePackageCount
}

// Collect implements required collect function for all promehteus collectors
func (collector *aptCollector) Collect(ch chan<- prometheus.Metric) {
	// Get labels for metrics
	hostname := os.Getenv("HOSTNAME")

	// Implement logic here to determine proper metric value to return to prometheus
	// for each descriptor or call other functions that do so.

	// Number of upgradable packages
	upgradablePackageCount, err := getUpgradablePackageCount()
	if err != nil {
		log.Fatal(err)
	}

	// Package cache timestamps
	packageCacheTimestamps, err := getPackageCacheTimestamps()
	if err != nil {
		log.Fatal(err)
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.packageCacheTimestamp, prometheus.GaugeValue, float64(packageCacheTimestamps), hostname)
	ch <- prometheus.MustNewConstMetric(collector.upgradablePackageCount, prometheus.GaugeValue, float64(upgradablePackageCount), hostname)
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
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("File '%s' does not exist.\n", packageCacheTimestampsFilePath)
	}

	return stats.ModTime().Unix(), nil
}
