package random

import (
	"fmt"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/metric/gauge"
	"math/rand/v2"
)

type Collector struct {
	min float64
	max float64
}

type ErrMinGreaterThanMax struct {
	Min float64
	Max float64
}

func (err ErrMinGreaterThanMax) Error() string {
	return fmt.Sprintf("Min=%g can`t be greater than Max=%g", err.Min, err.Max)
}

func newMinGreaterThanMaxError(min, max float64) ErrMinGreaterThanMax {
	return ErrMinGreaterThanMax{
		Min: min,
		Max: max,
	}
}

func New(min, max float64) (*Collector, error) {
	if max < min {
		return nil, newMinGreaterThanMaxError(min, max)
	}

	return &Collector{
		min: min,
		max: max,
	}, nil
}

func (collector *Collector) Collect() ([]_metric.Metric, error) {
	return []_metric.Metric{gauge.New(
		"RandomValue",
		// since rand.Float64 returns a value from 0 to 1 and does not support Min/Max
		// correct this by multiplying the value by the difference between Max and Min and adding Min to the result
		// the final value will be in the range from Min to Max including both values
		rand.Float64()*(collector.max-collector.min)+collector.min,
	)}, nil
}
