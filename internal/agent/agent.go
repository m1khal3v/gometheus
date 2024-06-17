package agent

import (
	"github.com/m1khal3v/gometheus/internal/agent/collector"
	"github.com/m1khal3v/gometheus/internal/agent/collector/random"
	"github.com/m1khal3v/gometheus/internal/agent/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/agent/storage"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/pkg/client"
	"time"
)

func Start(endpoint string, pollInterval, reportInterval uint32) {
	storage := storage.New()
	go collectMetrics(storage, pollInterval)
	sendMetrics(storage, client.New(endpoint), reportInterval)
}

func createCollectors() []collector.Collector {
	runtimeCollector := runtime.New()
	randomCollector, err := random.New(0, 512)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	return []collector.Collector{
		runtimeCollector,
		randomCollector,
	}
}

func collectMetrics(storage *storage.Storage, pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	collectors := createCollectors()

	for range ticker.C {
		for _, collector := range collectors {
			storage.Append(collector.Collect())
		}
	}
}

type metricSender interface {
	SendMetric(metricType, metricName, metricValue string) error
}

func sendMetrics(storage *storage.Storage, client metricSender, reportInterval uint32) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		storage.Remove(func(metric metric.Metric) bool {
			err := client.SendMetric(metric.GetType(), metric.GetName(), metric.GetStringValue())
			if err != nil {
				logger.Logger.Warn(err.Error())
				return false
			}

			return true
		})
	}
}
