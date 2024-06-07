package runtime

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/logger"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"reflect"
	"runtime"
)

type Collector struct {
	PollCount uint8
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

func NewCollector() *Collector {
	return &Collector{PollCount: 0}
}

func (collector *Collector) Collect() ([]*_metric.Metric, error) {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	metrics := make([]*_metric.Metric, 0, 28)
	for _, name := range getCollectableMemStatsMetrics() {
		metrics = append(metrics, collector.collectMetric(memStats, name))
	}
	metrics = append(metrics, collector.getPollCount())
	collector.refreshPollCount()

	return metrics, nil
}

func (collector *Collector) collectMetric(memStats *runtime.MemStats, name string) *_metric.Metric {
	field := reflect.ValueOf(*memStats).FieldByName(name)
	if !field.IsValid() {
		logger.Logger.Panic(fmt.Sprintf("Property '%v' not found in memStats", name))
	}

	metric, err := _metric.NewMetric(
		_metric.TypeGauge,
		name,
		field.Convert(reflect.TypeOf(float64(0))).Float(),
	)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	collector.PollCount = collector.PollCount + 1

	return metric
}

func (collector *Collector) getPollCount() *_metric.Metric {
	metric, err := _metric.NewMetric(
		_metric.TypeCounter,
		"PollCount",
		int64(collector.PollCount),
	)

	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	return metric
}

func (collector *Collector) refreshPollCount() {
	collector.PollCount = 0
}
