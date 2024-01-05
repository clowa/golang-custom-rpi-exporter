package exporter

import (
	"fmt"
	"os"

	"github.com/arduino/go-apt-client"
)

func GetUpgradablePackageCount() (int, error) {
	upgradablePackages, err := apt.ListUpgradable()
	if err != nil {
		return 0, err
	}

	return len(upgradablePackages), nil
}

func GetPackageCacheTimestamps() (int64, error) {
	const packageCacheTimestampsFilePath = "/var/lib/apt/lists"

	// Get File Stats
	stats, err := os.Stat(packageCacheTimestampsFilePath)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("File '%s' does not exist.\n", packageCacheTimestampsFilePath)
	}

	return stats.ModTime().Unix(), nil
}
