package counter

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/metric"
)

const Type = "counter"

type Metric struct {
	name  string
	value int64
}

func (metric *Metric) GetType() string {
	return Type
}

func (metric *Metric) GetName() string {
	return metric.name
}

func (metric *Metric) GetStringValue() string {
	return fmt.Sprintf("%d", metric.value)
}

func (metric *Metric) Replace(newMetric metric.Metric) metric.Metric {
	if metric.GetType() == newMetric.GetType() {
		newMetric.(*Metric).value += metric.value
	}

	return newMetric
}

func New(name string, value int64) *Metric {
	return &Metric{
		name:  name,
		value: value,
	}
}
