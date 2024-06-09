package gauge

import "fmt"

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

func (metric *Metric) GetValue() any {
	return metric.value
}

func (metric *Metric) String() string {
	return fmt.Sprintf("%f", metric.value)
}

func New(name string, value float64) *Metric {
	return &Metric{
		name:  name,
		value: value,
	}
}
