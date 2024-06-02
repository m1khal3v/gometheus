package collector

import "github.com/m1khal3v/gometheus/internal/storage"

type Collector interface {
	Collect() ([]*storage.Metric, error)
}

func CollectAll(collectors ...Collector) ([]*storage.Metric, error) {
	var allMetrics []*storage.Metric

	for _, collector := range collectors {
		metrics, err := collector.Collect()
		if err != nil {
			return nil, err
		}

		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}
