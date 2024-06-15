package agent

import (
	"github.com/m1khal3v/gometheus/internal/agent/storage"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/metric"
	"time"
)

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
