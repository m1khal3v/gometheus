package counter

import "fmt"

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

func (metric *Metric) Add(other *Metric) *Metric {
	metric.value += other.value

	return metric
}

func New(name string, value int64) *Metric {
	return &Metric{
		name:  name,
		value: value,
	}
}
