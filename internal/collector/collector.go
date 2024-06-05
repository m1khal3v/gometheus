package collector

import (
	"errors"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Collector interface {
	Collect() ([]*_metric.Metric, error)
}

func CollectAll(collectors ...Collector) ([]*_metric.Metric, error) {
	var allMetrics = make([]*_metric.Metric, 0)
	var allErrors error = nil

	for _, collector := range collectors {
		metrics, err := collector.Collect()
		if err != nil {
			allErrors = errors.Join(allErrors, err)
			continue
		}

		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, allErrors
}
