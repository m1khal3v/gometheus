// Package collector
// contains interface and its implementations
package collector

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
)

type Collector interface {
	Collect() (<-chan metric.Metric, error) // Collect metrics as channel(generator) from Collector
}
