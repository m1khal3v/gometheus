package runtime

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"reflect"
	"runtime"
)

type Collector struct {
	pollCount uint8
}

func getCollectableMemStatsMetrics() []string {
	return []string{
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
	}
}

func New() *Collector {
	return &Collector{pollCount: 0}
}

func (collector *Collector) Collect() []metric.Metric {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	metrics := make([]metric.Metric, 0, 28)
	for _, name := range getCollectableMemStatsMetrics() {
		metrics = append(metrics, collector.collectMetric(memStats, name))
	}
	metrics = append(metrics, collector.getPollCount())
	collector.refreshPollCount()

	return metrics
}

func (collector *Collector) collectMetric(memStats *runtime.MemStats, name string) metric.Metric {
	field := reflect.ValueOf(*memStats).FieldByName(name)
	if !field.IsValid() {
		logger.Logger.Panic(fmt.Sprintf("Property '%s' not found in memStats", name))
	}

	collector.pollCount = collector.pollCount + 1

	return gauge.New(
		name,
		field.Convert(reflect.TypeOf(float64(0))).Float(),
	)
}

func (collector *Collector) getPollCount() metric.Metric {
	return counter.New(
		"PollCount",
		int64(collector.pollCount),
	)
}

func (collector *Collector) refreshPollCount() {
	collector.pollCount = 0
}
