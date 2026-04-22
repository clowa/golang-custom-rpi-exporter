package cache

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Labels struct {
	Hostname string
}

type labelCache struct {
	mu           sync.RWMutex
	hostnameFunc func() string

	needsRefresh bool
	hostname     string
}

var globalLabelCache = newLabelCache(getHostname)

func newLabelCache(hostnameFunc func() string) *labelCache {
	return &labelCache{
		hostnameFunc: hostnameFunc,
		needsRefresh: true,
		hostname:     "",
	}
}

func GetLabelCache() *labelCache {
	return globalLabelCache
}

func (l *labelCache) NeedsRefresh() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.needsRefresh
}

func (l *labelCache) Refresh() {
	log.Info("Refreshing label cache")

	hostname := l.hostnameFunc()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.needsRefresh = false
	l.hostname = hostname
}

func (l *labelCache) GetLabels() Labels {
	l.mu.RLock()
	needsRefresh := l.needsRefresh
	hostname := l.hostname
	l.mu.RUnlock()

	if needsRefresh {
		l.Refresh()

		l.mu.RLock()
		hostname = l.hostname
		l.mu.RUnlock()
	}

	return Labels{Hostname: hostname}
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
