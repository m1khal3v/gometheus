package counter

import (
	"fmt"
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

func (metric *Metric) GetValue() int64 {
	return metric.value
}

func (metric *Metric) Add(value int64) {
	metric.value += value
}

func New(name string, value int64) *Metric {
	return &Metric{
		name:  name,
		value: value,
	}
}
