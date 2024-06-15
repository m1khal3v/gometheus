package agent

import (
	"github.com/m1khal3v/gometheus/internal/agent/storage"
	"github.com/m1khal3v/gometheus/pkg/client"
)

func Start(endpoint string, pollInterval, reportInterval uint32) {
	storage := storage.New()
	go collectMetrics(storage, pollInterval)
	sendMetrics(storage, client.New(endpoint), reportInterval)
}
