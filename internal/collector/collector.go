package collector

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Collector interface {
	Collect() []_metric.Metric
}

func CollectAll(collectors ...Collector) []_metric.Metric {
	var allMetrics = make([]_metric.Metric, 0)

	for _, collector := range collectors {
		metrics := collector.Collect()

		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics
}
