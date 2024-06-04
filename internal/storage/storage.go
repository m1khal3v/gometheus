package storage

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/store"
)

type Storage interface {
	Save(metric *store.Metric) error
	Get(name string) (*store.Metric, error)
	GetAll() (map[string]*store.Metric, error)
}

type MetricNotFoundError struct {
	Name string
}

func (err MetricNotFoundError) Error() string {
	return fmt.Sprintf("Metric '%v' not found", err.Name)
}

func NewMetricNotFoundError(name string) MetricNotFoundError {
	return MetricNotFoundError{
		Name: name,
	}
}
