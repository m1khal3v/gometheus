package collector

import "github.com/m1khal3v/gometheus/internal/metric"

type Collector interface {
	Collect() []metric.Metric
}
