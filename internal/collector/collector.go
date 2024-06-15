package collector

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Collector interface {
	Collect() []_metric.Metric
}
