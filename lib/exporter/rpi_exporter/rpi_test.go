package rpi_exporter

import (
	"errors"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRpiCollectorCollectErrorExportsScrapeSuccessZero(t *testing.T) {
	t.Parallel()

	collector := newRpiCollector()
	collector.getHostname = func() string { return "test-host" }
	collector.getRebootRequired = func() int { return 1 }
	collector.getCPUTemperature = func() (float64, error) { return 0, errors.New("temperature read failed") }

	registry := prometheus.NewRegistry()
	if err := registry.Register(collector); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	expected := strings.NewReader(`# HELP node_rpi_reboot_required Wether a Node reboot is required for software updates.
# TYPE node_rpi_reboot_required gauge
node_rpi_reboot_required{instance="test-host"} 1
# HELP node_rpi_scrape_success Whether the rpi collector scrape was successful (1 for success, 0 for failure).
# TYPE node_rpi_scrape_success gauge
node_rpi_scrape_success{instance="test-host"} 0
`)

	if err := testutil.GatherAndCompare(registry, expected, "node_rpi_reboot_required", "node_rpi_scrape_success"); err != nil {
		t.Fatalf("gather and compare: %v", err)
	}
}
