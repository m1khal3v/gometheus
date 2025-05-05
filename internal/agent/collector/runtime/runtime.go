// Package runtime
// collector for golang runtime package metrics
package runtime

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
)

type Collector struct {
	metrics []string
}

var ErrEmptyMetrics = errors.New("metrics are empty")
var float64Type = reflect.TypeOf(float64(0))

type InvalidMetricNameError struct {
	Name string
}

func (err InvalidMetricNameError) Error() string {
	return fmt.Sprintf("invalid metric name: %s", err.Name)
}

func newErrInvalidMetricName(name string) error {
	return &InvalidMetricNameError{
		Name: name,
	}
}

func New(metrics []string) (*Collector, error) {
	if len(metrics) == 0 {
		return nil, ErrEmptyMetrics
	}

	collector := &Collector{
		metrics: metrics,
	}

	if err := collector.validateMetricNames(); err != nil {
		return nil, err
	}

	return collector, nil
}

func (collector *Collector) validateMetricNames() error {
	var err error = nil
	value := reflect.ValueOf(runtime.MemStats{})

	for _, name := range collector.metrics {
		if !value.FieldByName(name).IsValid() {
			err = errors.Join(err, newErrInvalidMetricName(name))
		}
	}

	return err
}

func (collector *Collector) Collect() (<-chan metric.Metric, error) {
	channel := make(chan metric.Metric, len(collector.metrics)+1)

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	value := reflect.ValueOf(memStats)

	var pollCount int64
	var waitGroup sync.WaitGroup

	for _, name := range collector.metrics {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			channel <- gauge.New(
				name,
				value.FieldByName(name).Convert(float64Type).Float(),
			)

			atomic.AddInt64(&pollCount, 1)
		}()
	}

	waitGroup.Wait()
	channel <- counter.New(
		"PollCount",
		pollCount,
	)
	close(channel)

	return channel, nil
}
