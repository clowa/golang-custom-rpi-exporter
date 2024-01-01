package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"
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
	disableHttpCollectorFlag := flag.Bool("disable-http-collector", false, "Disables the default HTTP metrics endpoint")
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

	var wg sync.WaitGroup
	// Add a new goroutine to refresh metrics every 10 minutes
	wg.Add(1)
	go func() {
		// Decrement the counter when the goroutine completes.
		defer wg.Done()

		for true {
			log.Info("Refreshing metrics...")
			m.RefreshMetrics()
			if *enableTextfileCollectorFlag {
				log.Info("Writing metrics to file: ", metricFilePath)
				// Write latest metrics to file
				err := prometheus.WriteToTextfile(metricFilePath, reg)
				if err != nil {
					log.Fatal(err)
				}
			}
			time.Sleep(10 * time.Minute)
		}
	}()

	if !*disableHttpCollectorFlag {
		// Expose metrics and custom registry via an HTTP server.
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

	// Wait for all goroutines to complete
	wg.Wait()
}
