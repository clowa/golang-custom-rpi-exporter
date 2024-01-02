package metrics

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type metrics struct {
	CpuTemp                   *prometheus.GaugeVec
	RebootRequired            *prometheus.GaugeVec
	AptPackageCacheTimestamps *prometheus.GaugeVec
	AptUpgradablePackageCount *prometheus.GaugeVec
}

// Register all metrics with the given prometheus.Registerer.
func (m *metrics) RegisterAll(reg prometheus.Registerer) {
	log.Info("New Prometheus metric registered: ", "cpu_temperature_celsius")
	reg.MustRegister(m.CpuTemp)

	log.Info("New Prometheus metric registered: ", "reboot_required")
	reg.MustRegister(m.RebootRequired)

	log.Info("New Prometheus metric registered: ", "upgradable_packages")
	reg.MustRegister(m.AptUpgradablePackageCount)

	log.Info("New Prometheus metric registered: ", "package_cache_timestamp_seconds")
	reg.MustRegister(m.AptPackageCacheTimestamps)
}

// Init initializes all Prometheus metrics made available by this exporter.
func Init(reg prometheus.Registerer) *metrics {
	m := &metrics{
		CpuTemp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "node",
			Name:      "cpu_temperature_celsius",
			Help:      "Current temperature of the CPU in degrees Celsius.",
		},
			[]string{"instance"},
		),
		RebootRequired: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "node",
				Name:      "reboot_required",
				Help:      "Wether a Node reboot is required for software updates.",
			},
			[]string{"instance"},
		),
		AptUpgradablePackageCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "apt",
				Name:      "upgradable_packages",
				Help:      "Number of upgradable packages.",
			},
			[]string{"instance"},
		),
		AptPackageCacheTimestamps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "apt",
				Name:      "package_cache_timestamp_seconds",
				Help:      "Unix timestamp of the package cache in seconds.",
			},
			[]string{"instance"},
		),
	}

	m.RegisterAll(reg)
	return m
}

func (m *metrics) RefreshMetrics() {
	hostname := os.Getenv("HOSTNAME")
	// Set CPU temperature
	cpuTemp, err := GetTemperature()
	if err != nil {
		log.Fatal(err)
	}
	m.CpuTemp.WithLabelValues(hostname).Set(cpuTemp)

	// Set reboot required
	rebootRequired := GetRebootRequired()
	m.RebootRequired.WithLabelValues(hostname).Set(float64(rebootRequired))

	// Set number of upgradable packages
	upgradablePackageCount, err := GetUpgradablePackageCount()
	if err != nil {
		log.Fatal(err)
	}
	m.AptUpgradablePackageCount.WithLabelValues(hostname).Set(float64(upgradablePackageCount))

	// Get package cache timestamps
	packageCacheTimestamps, err := GetPackageCacheTimestamps()
	if err != nil {
		log.Fatal(err)
	}
	m.AptPackageCacheTimestamps.WithLabelValues(hostname).Set(float64(packageCacheTimestamps))
}
