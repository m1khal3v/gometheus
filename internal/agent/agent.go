package agent

import (
	"github.com/m1khal3v/gometheus/internal/collector"
	"github.com/m1khal3v/gometheus/internal/collector/random"
	"github.com/m1khal3v/gometheus/internal/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/logger"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/pkg/client"
	"time"
)

func Start(endpoint string, pollInterval, reportInterval uint32) {
	storage := newStorage()
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

func collectMetrics(storage *storage, pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	collectors := createCollectors()

	for range ticker.C {
		for _, collector := range collectors {
			storage.appendMetrics(collector.Collect())
		}
	}
}

func sendMetrics(storage *storage, client metricSender, reportInterval uint32) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		storage.removeMetrics(func(metric _metric.Metric) bool {
			err := client.SendMetric(metric.GetType(), metric.GetName(), metric.GetStringValue())
			if err != nil {
				logger.Logger.Warn(err.Error())
				return false
			}

			return true
		})
	}
}
