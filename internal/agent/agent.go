package agent

import (
	"github.com/m1khal3v/gometheus/internal/agent/collector"
	"github.com/m1khal3v/gometheus/internal/agent/collector/random"
	"github.com/m1khal3v/gometheus/internal/agent/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/agent/storage"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/pkg/client"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"time"
)

func Start(endpoint string, pollInterval, reportInterval uint32) {
	storage := storage.New()
	go collectMetrics(storage, pollInterval)
	saveMetrics(storage, client.New(endpoint, true), reportInterval)
}

func createCollectors() []collector.Collector {
	runtimeCollector := runtime.New()
	randomCollector, err := random.New(0, 512)
	if err != nil {
		logger.Logger.Fatal(err.Error())
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

type apiClient interface {
	SaveMetricAsJSON(request *request.SaveMetricRequest) (*response.SaveMetricResponse, error)
}

func saveMetrics(storage *storage.Storage, client apiClient, reportInterval uint32) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		storage.Remove(func(metric metric.Metric) bool {
			request, err := transformer.TransformToSaveRequest(metric)
			if err != nil {
				logger.Logger.Error(err.Error())
				return false
			}

			if _, err = client.SaveMetricAsJSON(request); err != nil {
				logger.Logger.Warn(err.Error())
				return false
			}

			return true
		})
	}
}
