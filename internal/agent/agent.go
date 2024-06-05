package agent

import (
	"github.com/m1khal3v/gometheus/internal/client"
	"github.com/m1khal3v/gometheus/internal/collector"
	"github.com/m1khal3v/gometheus/internal/collector/random"
	"github.com/m1khal3v/gometheus/internal/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/store"
	"sync"
	"time"
)

var mutex sync.Mutex
var allMetrics = make([]*store.Metric, 0)

func Start(endpoint string, pollInterval, reportInterval uint32) {
	go collectMetrics(pollInterval)
	sendMetrics(endpoint, reportInterval)
}

func collectMetrics(pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	runtimeCollector := runtime.NewCollector()
	randomCollector, err := random.NewCollector(0, 512)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	for range ticker.C {
		metrics, err := collector.CollectAll(
			runtimeCollector,
			randomCollector,
		)
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		mutex.Lock()
		allMetrics = append(allMetrics, metrics...)
		mutex.Unlock()
	}
}

func sendMetrics(endpoint string, reportInterval uint32) {
	apiClient := client.NewClient(endpoint)
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		mutex.Lock()
		retryMetrics := make([]*store.Metric, 0)
		for _, metric := range allMetrics {
			err := apiClient.SendMetric(metric)
			if err != nil {
				logger.Logger.Warn(err.Error())
				retryMetrics = append(retryMetrics, metric)
			}
		}
		allMetrics = retryMetrics
		mutex.Unlock()
	}
}
