package runtime

import (
	"errors"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"reflect"
	"runtime"
)

type Collector struct {
	pollCount uint8
	metrics   []string
	memStats  *runtime.MemStats
}

type ErrInvalidMetricName struct {
	Name string
}

func (err ErrInvalidMetricName) Error() string {
	return fmt.Sprintf("invalid metric name: %s", err.Name)
}

func newErrInvalidMetricName(name string) error {
	return &ErrInvalidMetricName{
		Name: name,
	}
}

func New(metrics []string) (*Collector, error) {
	collector := &Collector{
		pollCount: 0,
		metrics:   metrics,
		memStats:  &runtime.MemStats{},
	}

	if err := collector.validateMetricNames(); err != nil {
		return nil, err
	}

	return collector, nil
}

func (collector *Collector) validateMetricNames() error {
	var err error = nil

	for _, name := range collector.metrics {
		if !reflect.ValueOf(*collector.memStats).FieldByName(name).IsValid() {
			err = errors.Join(err, newErrInvalidMetricName(name))
		}
	}

	return err
}

func (collector *Collector) Collect() []metric.Metric {
	runtime.ReadMemStats(collector.memStats)
	metrics := make([]metric.Metric, 0)
	for _, name := range collector.metrics {
		metrics = append(metrics, collector.collectMetric(name))
	}
	metrics = append(metrics, collector.getPollCount())
	collector.refreshPollCount()

	return metrics
}

func (collector *Collector) collectMetric(name string) metric.Metric {
	field := reflect.ValueOf(*collector.memStats).FieldByName(name)
	collector.pollCount = collector.pollCount + 1

	return gauge.New(
		name,
		field.Convert(reflect.TypeOf(float64(0))).Float(),
	)
}

func (collector *Collector) getPollCount() metric.Metric {
	return counter.New(
		"PollCount",
		int64(collector.pollCount),
	)
}

func (collector *Collector) refreshPollCount() {
	collector.pollCount = 0
}
