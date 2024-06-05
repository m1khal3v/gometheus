package random

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/store"
	"math/rand/v2"
)

type Collector struct {
	Min float64
	Max float64
}

type MinGreaterThanMaxError struct {
	Min float64
	Max float64
}

func (err MinGreaterThanMaxError) Error() string {
	return fmt.Sprintf("Min=%v can`t be greater than Max=%v", err.Min, err.Max)
}

func newMinGreaterThanMaxError(min float64, max float64) MinGreaterThanMaxError {
	return MinGreaterThanMaxError{
		Min: min,
		Max: max,
	}
}

func NewCollector(min float64, max float64) (*Collector, error) {
	if max < min {
		return nil, newMinGreaterThanMaxError(min, max)
	}

	return &Collector{
		Min: min,
		Max: max,
	}, nil
}

func (collector *Collector) Collect() ([]*store.Metric, error) {
	metric, err := store.NewMetric(
		store.MetricTypeGauge,
		"RandomValue",
		rand.Float64()*(collector.Max-collector.Min)+collector.Min,
	)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	return []*store.Metric{metric}, nil
}
