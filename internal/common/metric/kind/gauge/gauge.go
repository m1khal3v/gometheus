package gauge

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
)

const MetricType = "gauge"

type Metric struct {
	name  string
	value float64
}

func (metric *Metric) Type() string {
	return MetricType
}

func (metric *Metric) Name() string {
	return metric.name
}

func (metric *Metric) StringValue() string {
	return fmt.Sprintf("%g", metric.value)
}

func (metric *Metric) GetValue() float64 {
	return metric.value
}

func (metric *Metric) Clone() metric.Metric {
	clone := *metric
	return &clone
}

func New(name string, value float64) *Metric {
	return &Metric{
		name:  name,
		value: value,
	}
}
