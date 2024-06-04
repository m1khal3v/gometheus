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

const pollInterval = 2 * time.Second
const reportInterval = 10 * time.Second

var mutex sync.Mutex
var allMetrics = make([]*store.Metric, 0)
var apiClient = client.NewClient()

func Start() {
	go collectMetrics()
	go sendMetrics()
	for {
		time.Sleep(1 * time.Second)
	}
}

func collectMetrics() {
	ticker := time.NewTicker(pollInterval)
	runtimeCollector := runtime.NewCollector()
	randomCollector, err := random.NewCollector(0, 512)
	if err != nil {
		panic(err)
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

func sendMetrics() {
	ticker := time.NewTicker(reportInterval)
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
