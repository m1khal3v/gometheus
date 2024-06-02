package random

import (
	"github.com/m1khal3v/gometheus/internal/storage"
	"math/rand/v2"
)

type Collector struct {
	Min float64
	Max float64
}

func NewCollector(min float64, max float64) *Collector {
	return &Collector{
		Min: min,
		Max: max,
	}
}

func (collector *Collector) Collect() ([]*storage.Metric, error) {
	metric, err := storage.NewMetric(
		storage.MetricTypeGauge,
		"RandomValue",
		rand.Float64()*(collector.Max-collector.Min)+collector.Min,
	)
	if err != nil {
		panic("Can't create 'RandomValue' metric")
	}

	return []*storage.Metric{metric}, nil
}
