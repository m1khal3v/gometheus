package collector

import (
	"errors"
	"github.com/m1khal3v/gometheus/internal/store"
)

type Collector interface {
	Collect() ([]*store.Metric, error)
}

func CollectAll(collectors ...Collector) ([]*store.Metric, error) {
	var allMetrics = make([]*store.Metric, 0)
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
