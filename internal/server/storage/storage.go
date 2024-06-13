package storage

import (
	"github.com/m1khal3v/gometheus/internal/metric"
)

type Storage interface {
	Save(metric metric.Metric)
	Get(name string) metric.Metric
	GetAll() map[string]metric.Metric
}
