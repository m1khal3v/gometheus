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
	"go.uber.org/zap"
	"time"
)

func createCollectors() ([]collector.Collector, error) {
	runtimeCollector, err := runtime.New([]string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	})
	if err != nil {
		return nil, err
	}

	randomCollector, err := random.New(0, 512)
	if err != nil {
		return nil, err
	}

	return []collector.Collector{
		runtimeCollector,
		randomCollector,
	}, nil
}

func Start(endpoint string, pollInterval, reportInterval uint32, batchSize uint64) error {
	storage := storage.New()
	client := client.New(endpoint, true)
	collectors, err := createCollectors()
	if err != nil {
		return err
	}

	go collectMetrics(storage, collectors, pollInterval)
	saveMetrics(storage, client, reportInterval, batchSize)

	return nil
}

func collectMetrics(storage *storage.Storage, collectors []collector.Collector, pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	for range ticker.C {
		for _, collector := range collectors {
			storage.Append(collector.Collect())
		}
	}
}

type apiClient interface {
	SaveMetricsAsJSON(requests []*request.SaveMetricRequest) ([]*response.SaveMetricResponse, *response.ApiError, error)
}

func saveMetrics(storage *storage.Storage, client apiClient, reportInterval uint32, batchSize uint64) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		storage.RemoveBatch(func(metrics []metric.Metric) bool {
			requests := make([]*request.SaveMetricRequest, 0, len(metrics))
			for _, metric := range metrics {
				request, err := transformer.TransformToSaveRequest(metric)
				if err != nil {
					logger.Logger.Error(err.Error())
					continue
				}

				requests = append(requests, request)
			}

			if _, apiErr, err := client.SaveMetricsAsJSON(requests); err != nil {
				logger.Logger.Warn(err.Error())
				if apiErr != nil {
					logger.Logger.Warn(
						apiErr.Message,
						zap.Int("code", apiErr.Code),
						zap.Any("details", apiErr.Details),
					)
				}
				return false
			}

			return true
		}, batchSize)
	}
}
