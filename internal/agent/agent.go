package agent

import (
	"github.com/m1khal3v/gometheus/internal/collector"
	"github.com/m1khal3v/gometheus/internal/collector/random"
	"github.com/m1khal3v/gometheus/internal/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/logger"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/pkg/client"
	"sync"
	"time"
)

var mutex sync.Mutex
var allMetrics = make([]_metric.Metric, 0)

func Start(endpoint string, pollInterval, reportInterval uint32) {
	go collectMetrics(pollInterval)
	sendMetrics(client.New(endpoint), reportInterval)
}

func collectMetrics(pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	runtimeCollector := runtime.New()
	randomCollector, err := random.New(0, 512)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	for range ticker.C {
		metrics := collector.CollectAll(
			runtimeCollector,
			randomCollector,
		)
		mutex.Lock()
		allMetrics = append(allMetrics, metrics...)
		mutex.Unlock()
	}
}

func sendMetrics(client metricSender, reportInterval uint32) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		mutex.Lock()
		retryMetrics := make([]_metric.Metric, 0)
		for _, metric := range allMetrics {
			err := client.SendMetric(metric.GetType(), metric.GetName(), metric.GetStringValue())
			if err != nil {
				logger.Logger.Warn(err.Error())
				retryMetrics = append(retryMetrics, metric)
			}
		}
		allMetrics = retryMetrics
		mutex.Unlock()
	}
}
