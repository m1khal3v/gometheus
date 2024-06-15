package agent

import (
	"github.com/m1khal3v/gometheus/internal/agent/collector"
	"github.com/m1khal3v/gometheus/internal/agent/collector/random"
	"github.com/m1khal3v/gometheus/internal/agent/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/agent/storage"
	"github.com/m1khal3v/gometheus/internal/logger"
	"time"
)

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
