package memory

import (
	metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Storage struct {
	metrics map[string]metric.Metric
}

func (storage *Storage) Get(name string) metric.Metric {
	return storage.metrics[name]
}

func (storage *Storage) GetAll() map[string]metric.Metric {
	return storage.metrics
}

func (storage *Storage) Save(metric metric.Metric) {
	storage.metrics[metric.GetName()] = metric
}

func New() *Storage {
	return &Storage{metrics: make(map[string]metric.Metric)}
}
