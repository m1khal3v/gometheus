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
}

var ErrEmptyMetrics = errors.New("metrics are empty")

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
	if len(metrics) == 0 {
		return nil, ErrEmptyMetrics
	}

	collector := &Collector{
		pollCount: 0,
		metrics:   metrics,
	}

	if err := collector.validateMetricNames(); err != nil {
		return nil, err
	}

	return collector, nil
}

func (collector *Collector) validateMetricNames() error {
	var err error = nil
	memStats := runtime.MemStats{}

	for _, name := range collector.metrics {
		if !reflect.ValueOf(memStats).FieldByName(name).IsValid() {
			err = errors.Join(err, newErrInvalidMetricName(name))
		}
	}

	return err
}

func (collector *Collector) Collect() []metric.Metric {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	metrics := make([]metric.Metric, 0)
	for _, name := range collector.metrics {
		metrics = append(metrics, collector.collectMetric(memStats, name))
	}
	metrics = append(metrics, collector.getPollCount())

	collector.refreshPollCount()

	return metrics
}

func (collector *Collector) collectMetric(memStats runtime.MemStats, name string) metric.Metric {
	field := reflect.ValueOf(memStats).FieldByName(name)
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
