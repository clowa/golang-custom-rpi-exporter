package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/clowa/golang-custom-rpi-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	metricFolder = "/var/lib/prometheus/node-exporter"
)

func createDirectoryIfNotExists(path string, mod fs.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("Creating directory: ", path)
		err = os.MkdirAll(path, mod)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	const metricFilePath = metricFolder + "/custom_node_metrics.prom"

	// Command line flags
	enableTextfileCollectorFlag := flag.Bool("enable-textfile-collector", false, fmt.Sprintf("Exports metrics additionally as .prom file to %s", metricFilePath))

	flag.Parse()

	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	m := metrics.Init(reg)

	if *enableTextfileCollectorFlag {
		err := createDirectoryIfNotExists(metricFolder, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		for true {
			log.Info("Refreshing metrics...")
			m.RefreshMetrics()
			if *enableTextfileCollectorFlag {
				log.Info("Writing metrics to file: ", metricFilePath)
				// Write latest metrics to file
				prometheus.WriteToTextfile(metricFilePath, reg)
			}
			time.Sleep(10 * time.Minute)
		}
	}()

	// Expose metrics and custom registry via an HTTP server.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
