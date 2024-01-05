package exporter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Check if the reboot-required file exists
func GetRebootRequired() int {
	const rebootRequiredFilePath = "/var/run/reboot-required"

	// Check if the file exists
	if _, err := os.Stat(rebootRequiredFilePath); os.IsNotExist(err) {
		return 0
	}

	return 1
}

// Get the CPU temperature in degrees Celsius from /sys/class/thermal/thermal_zone0/temp
func GetTemperature() (float64, error) {
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
