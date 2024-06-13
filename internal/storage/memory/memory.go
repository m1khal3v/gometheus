package memory

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
)

type Storage struct {
	metrics map[string]_metric.Metric
}

func (storage *Storage) Get(name string) _metric.Metric {
	return storage.metrics[name]
}

func (storage *Storage) GetAll() map[string]_metric.Metric {
	return storage.metrics
}

func (storage *Storage) Save(metric _metric.Metric) {
	current, ok := storage.metrics[metric.GetName()]
	if !ok {
		storage.metrics[metric.GetName()] = metric
	} else {
		storage.metrics[metric.GetName()] = _metric.Combine(current, metric)
	}
}

func New() *Storage {
	return &Storage{metrics: make(map[string]_metric.Metric)}
}
