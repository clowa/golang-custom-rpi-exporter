package cache

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type labelCache struct {
	needsRefresh bool
	Hostname     string
}

var globalLabelCache = labelCache{
	needsRefresh: true,
	Hostname:     "",
}

func GetLabelCache() *labelCache {
	return &globalLabelCache
}

func (l *labelCache) NeedsRefresh() bool {
	return globalLabelCache.needsRefresh
}

func (l *labelCache) Refresh() {
	log.Info("Refreshing label cache")
	l.needsRefresh = false
	l.Hostname = getHostname()
}

func getHostname() string {
	var err error

	hostname := os.Getenv("HOSTNAME")

	if hostname == "" {
		log.Info("HOSTNAME environment variable not set, attempting to determine hostname from OS")
		hostname, err = os.Hostname()
		if err != nil {
			log.Error("Unable to determine hostname")
			hostname = "unknown"
		}
	}
	return hostname
}
