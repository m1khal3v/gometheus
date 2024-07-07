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

func Start(endpoint string, pollInterval, reportInterval uint32) error {
	storage := storage.New()
	client := client.New(endpoint, true)
	collectors, err := createCollectors()
	if err != nil {
		return err
	}

	go collectMetrics(storage, collectors, pollInterval)
	saveMetrics(storage, client, reportInterval)

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
	SaveMetricAsJSON(request *request.SaveMetricRequest) (*response.SaveMetricResponse, error)
}

func saveMetrics(storage *storage.Storage, client apiClient, reportInterval uint32) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for range ticker.C {
		storage.Remove(func(metric metric.Metric) bool {
			request, err := transformer.TransformToSaveRequest(metric)
			if err != nil {
				logger.Logger.Error(err.Error())
				return true
			}

			if _, err = client.SaveMetricAsJSON(request); err != nil {
				logger.Logger.Warn(err.Error())
				return false
			}

			return true
		})
	}
}
