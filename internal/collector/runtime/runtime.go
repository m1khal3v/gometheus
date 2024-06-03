package runtime

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/storage"
	"reflect"
	"runtime"
	"sync"
)

type Collector struct {
	PollCount uint8
}

var waitGroup sync.WaitGroup

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

func (collector *Collector) Collect() ([]*storage.Metric, error) {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	metrics := make([]*storage.Metric, 29)
	for _, name := range getCollectableMemStatsMetrics() {
		waitGroup.Add(1)
		go collector.collectMetric(metrics, memStats, name)
	}
	waitGroup.Wait()
	metrics = append(metrics, collector.getPollCount())
	collector.refreshPollCount()

	return metrics, nil
}

func (collector *Collector) collectMetric(metrics []*storage.Metric, memStats *runtime.MemStats, name string) {
	field := reflect.ValueOf(*memStats).FieldByName(name)
	if !field.IsValid() {
		panic(fmt.Sprintf("Property '%v' not found in memStats", name))
	}

	metric, err := storage.NewMetric(
		storage.MetricTypeGauge,
		name,
		field.Convert(reflect.TypeOf(float64(0))).Float(),
	)
	if err != nil {
		panic(fmt.Sprintf("Can't create '%v' metric", name))
	}

	metrics = append(metrics, metric)
	collector.PollCount = collector.PollCount + 1
	waitGroup.Done()
}

func (collector *Collector) getPollCount() *storage.Metric {
	metric, err := storage.NewMetric(
		storage.MetricTypeCounter,
		"PollCount",
		int64(collector.PollCount),
	)

	if err != nil {
		panic("Can't create PollCount metric")
	}

	return metric
}

func (collector *Collector) refreshPollCount() {
	collector.PollCount = 0
}
