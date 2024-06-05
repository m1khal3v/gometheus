package storage

import (
	"fmt"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Storage interface {
	Save(metric *_metric.Metric) error
	Get(name string) (*_metric.Metric, error)
	GetAll() (map[string]_metric.Metric, error)
}

type ErrMetricNotFound struct {
	Name string
}

func (err ErrMetricNotFound) Error() string {
	return fmt.Sprintf("Metric '%v' not found", err.Name)
}

func NewMetricNotFoundError(name string) ErrMetricNotFound {
	return ErrMetricNotFound{
		Name: name,
	}
}
