package rpi_exporter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/clowa/golang-custom-rpi-exporter/lib/cache"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type rpiCollector struct {
	rebootRequired *prometheus.Desc
	cpuTemperature *prometheus.Desc
	scrapeSuccess  *prometheus.Desc

	getHostname       func() string
	getRebootRequired func() int
	getCPUTemperature func() (float64, error)
}

func Register() {
	collector := newRpiCollector()
	prometheus.MustRegister(collector)
}

func newRpiCollector() *rpiCollector {
	const namespace = "node"

	return &rpiCollector{
		rebootRequired: prometheus.NewDesc(prometheus.BuildFQName(namespace, "rpi", "reboot_required"),
			"Wether a Node reboot is required for software updates.",
			[]string{"instance"}, nil,
		),
		cpuTemperature: prometheus.NewDesc(prometheus.BuildFQName(namespace, "rpi", "cpu_temperature_celsius"),
			"Current temperature of the CPU in degrees Celsius.",
			[]string{"instance"}, nil,
		),
		scrapeSuccess: prometheus.NewDesc(prometheus.BuildFQName(namespace, "rpi", "scrape_success"),
			"Whether the rpi collector scrape was successful (1 for success, 0 for failure).",
			[]string{"instance"}, nil,
		),
		getHostname:       func() string { return cache.GetLabelCache().GetLabels().Hostname },
		getRebootRequired: getRebootRequired,
		getCPUTemperature: getCpuTemperature,
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *rpiCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.rebootRequired
	ch <- collector.cpuTemperature
	ch <- collector.scrapeSuccess
}

// Collect implements required collect function for all promehteus collectors
func (collector *rpiCollector) Collect(ch chan<- prometheus.Metric) {
	instance := collector.getHostname()
	scrapeSuccess := 1.0

	rebootRequired := collector.getRebootRequired()
	cpuTemp, err := collector.getCPUTemperature()
	if err != nil {
		scrapeSuccess = 0
		log.WithError(err).Error("rpi collector scrape failed")
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.cpuTemperature, prometheus.GaugeValue, float64(cpuTemp), instance)
	}
	ch <- prometheus.MustNewConstMetric(collector.rebootRequired, prometheus.GaugeValue, float64(rebootRequired), instance)
	ch <- prometheus.MustNewConstMetric(collector.scrapeSuccess, prometheus.GaugeValue, scrapeSuccess, instance)
}

// getRebootRequired returns 1 if a reboot is required for software updates, 0 otherwise
func getRebootRequired() int {
	const rebootRequiredFilePath = "/var/run/reboot-required"

	// Check if the file exists
	if _, err := os.Stat(rebootRequiredFilePath); os.IsNotExist(err) {
		return 0
	}

	return 1
}

// getCpuTemperature returns the CPU temperature in degrees Celsius
func getCpuTemperature() (float64, error) {
	const temperatureFilePath = "/sys/class/thermal/thermal_zone0/temp"

	// Check if the file exists
	if _, err := os.Stat(temperatureFilePath); err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("temperature file %q does not exist: %w", temperatureFilePath, err)
		}

		return 0, fmt.Errorf("failed to stat temperature file %q: %w", temperatureFilePath, err)
	}

	// Read the file content
	content, err := os.ReadFile(temperatureFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read temperature file %q: %w", temperatureFilePath, err)
	}

	// Convert file content to string and remove trailing newline characters
	fileContent := strings.TrimSpace(string(content))

	// Convert string to integer
	milliDegrees, err := strconv.Atoi(fileContent)

	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature value %q: %w", fileContent, err)
	}

	// Convert millidegrees to degrees (format as a string)
	degrees := float64(milliDegrees) / 1000.0
	return degrees, nil
}
