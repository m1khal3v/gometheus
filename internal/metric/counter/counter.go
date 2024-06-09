package counter

import "fmt"

const Type = "counter"

type ErrNamesDontMatch struct {
	Name1, Name2 string
}

func (e ErrNamesDontMatch) Error() string {
	return fmt.Sprintf("name %s and name %s don't match", e.Name1, e.Name2)
}

func newErrNamesDontMatch(name1, name2 string) error {
	return ErrNamesDontMatch{
		Name1: name1,
		Name2: name2,
	}
}

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

func (metric *Metric) GetValue() any {
	return metric.value
}

func (metric *Metric) String() string {
	return fmt.Sprintf("%v", metric.value)
}

func (metric *Metric) Add(other *Metric) (*Metric, error) {
	if metric.name != other.name {
		return metric, newErrNamesDontMatch(metric.name, other.name)
	}

	metric.value += other.value

	return metric, nil
}

func New(name string, value int64) *Metric {
	return &Metric{
		name:  name,
		value: value,
	}
}
