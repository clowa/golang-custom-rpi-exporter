package main

import (
	"flag"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/clowa/golang-custom-rpi-exporter/exporter/apt_exporter"
	"github.com/clowa/golang-custom-rpi-exporter/exporter/rpi_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	exporterDisplayName = "Custom Raspberry Pi Exporter"
	exporterName        = "custom_rpi_exporter"
)

// Structure to hold the configuration of the exporter
type config struct {
	listenAddress               *string
	metricPath                  *string
	httpCollectorDisabledFlag   *bool
	textFileExporterEnabledFlag *bool
	textFileExporterDestination *string
}

func main() {
	config := readFlags()

	if *config.textFileExporterEnabledFlag {
		metricFolder := filepath.Dir(*config.textFileExporterDestination)
		err := createDirectoryIfNotExists(metricFolder, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Register collectors
	apt_exporter.Register()
	rpi_exporter.Register()

	// Add a new goroutine to refresh metrics every 10 minutes
	var wg sync.WaitGroup

	if *config.textFileExporterEnabledFlag {
		spawnFileWriter(&wg, *config.textFileExporterDestination, 1*time.Minute)
	}

	if !*config.httpCollectorDisabledFlag {
		// Expose metrics and custom registry via an HTTP server.
		log.Fatal(serverMetrics(*config.listenAddress, *config.metricPath))
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

// readFlags reads the command line flags and returns a config struct.
func readFlags() *config {
	var config config

	config.listenAddress = flag.String("address", ":8080", "Address on which to expose metrics.")
	config.metricPath = flag.String("path", "/metrics", "Path under which to expose metrics.")

	config.httpCollectorDisabledFlag = flag.Bool("disable-http-collector", false, "Disables the default HTTP metrics endpoint")

	config.textFileExporterEnabledFlag = flag.Bool("enable-textfile-collector", false, "Export metrics additionally to file system")
	config.textFileExporterDestination = flag.String("textfile-collector-destination", "/var/lib/prometheus/node-exporter/"+exporterName+".prom", "Path to write metrics as file to.")

	flag.Parse()
	return &config
}

// createDirectoryIfNotExists creates a directory with a certain mode if it does not exist.
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

// spawnFileWriter spawns a goroutine that writes the latest metrics to a text file exporter on a given interval.
func spawnFileWriter(wg *sync.WaitGroup, filePath string, interval time.Duration) {
	wg.Add(1)

	go func() {
		for {
			log.Info("Writing metrics to file: ", filePath)
			// Write latest metrics to file
			err := prometheus.WriteToTextfile(filePath, prometheus.DefaultGatherer)
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(interval)
		}
	}()
}

// serverMetrics exposes the metrics endpoint via HTTP.
func serverMetrics(listenAddress, metricsPath string) error {
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>` + exporterDisplayName + `</title></head>
			<body>
			<h1>` + exporterDisplayName + `</h1>
			<p><a href='` + metricsPath + `'>Metrics</a></p>
			</body>
			</html>
		`))
	})

	log.Printf("Starting Server on Port %s", listenAddress)
	return http.ListenAndServe(listenAddress, nil)
}
