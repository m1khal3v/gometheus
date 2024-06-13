package storage

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Storage interface {
	Save(metric _metric.Metric)
	Get(name string) _metric.Metric
	GetAll() map[string]_metric.Metric
}
