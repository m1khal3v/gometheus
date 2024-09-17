// Package random
// collector for random value metric
package random

import (
	"fmt"
	"math/rand/v2"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
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
	return fmt.Sprintf("min=%g can`t be greater than max=%g", err.Min, err.Max)
}

func newErrMinGreaterThanMax(min, max float64) error {
	return &ErrMinGreaterThanMax{
		Min: min,
		Max: max,
	}
}

func New(min, max float64) (*Collector, error) {
	if max < min {
		return nil, newErrMinGreaterThanMax(min, max)
	}

	return &Collector{
		min: min,
		max: max,
	}, nil
}

func (collector *Collector) Collect() (<-chan metric.Metric, error) {
	channel := make(chan metric.Metric, 1)

	go func() {
		channel <- gauge.New(
			"RandomValue",
			// since rand.Float64 returns a value from 0 to 1 and does not support Min/Max
			// correct this by multiplying the value by the difference between Max and Min and adding Min to the result
			// the final value will be in the range from Min to Max including both values
			rand.Float64()*(collector.max-collector.min)+collector.min,
		)
		close(channel)
	}()

	return channel, nil
}
