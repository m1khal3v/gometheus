package gauge

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/metric"
)

const Type = "gauge"

type Metric struct {
	name  string
	value float64
}

func (metric *Metric) GetType() string {
	return Type
}

func (metric *Metric) GetName() string {
	return metric.name
}

func (metric *Metric) GetStringValue() string {
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
