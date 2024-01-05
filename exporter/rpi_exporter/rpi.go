package rpi_exporter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type rpiCollector struct {
	rebootRequired *prometheus.Desc
	cpuTemperature *prometheus.Desc
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
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *rpiCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.rebootRequired
	ch <- collector.cpuTemperature
}

// Collect implements required collect function for all promehteus collectors
func (collector *rpiCollector) Collect(ch chan<- prometheus.Metric) {
	// Get labels for metrics
	hostname := os.Getenv("HOSTNAME")

	// Implement logic here to determine proper metric value to return to prometheus
	// for each descriptor or call other functions that do so.

	// Reboot required
	rebootRequired := getRebootRequired()

	// CPU temperature
	cpuTemp, err := getCpuTemperature()
	if err != nil {
		log.Fatal(err)
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.cpuTemperature, prometheus.GaugeValue, float64(cpuTemp), hostname)
	ch <- prometheus.MustNewConstMetric(collector.rebootRequired, prometheus.GaugeValue, float64(rebootRequired), hostname)
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
	if _, err := os.Stat(temperatureFilePath); os.IsNotExist(err) {
		return 0, fmt.Errorf("File '%s' does not exist.\n", temperatureFilePath)
	}

	// Read the file content
	content, err := ioutil.ReadFile(temperatureFilePath)
	if err != nil {
		return 0, fmt.Errorf("Error reading file: %s\n", err)
	}

	// Convert file content to string and remove trailing newline characters
	fileContent := strings.TrimSpace(string(content))

	// Convert string to integer
	milliDegrees, err := strconv.Atoi(fileContent)

	if err != nil {
		return 0, fmt.Errorf("Error converting string to integer: %s\n", err)
	}

	// Convert millidegrees to degrees (format as a string)
	degrees := float64(milliDegrees) / 1000.0
	return degrees, nil
}
