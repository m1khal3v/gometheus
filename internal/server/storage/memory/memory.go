package memory

import (
	"github.com/m1khal3v/gometheus/internal/metric"
)

type Storage struct {
	metrics map[string]metric.Metric
}

func (storage *Storage) Get(name string) metric.Metric {
	metric, ok := storage.metrics[name]
	if !ok {
		return nil
	}

	return metric.Clone()
}

func (storage *Storage) GetAll() map[string]metric.Metric {
	clone := make(map[string]metric.Metric)
	for name, metric := range storage.metrics {
		clone[name] = metric.Clone()
	}

	return clone
}

func (storage *Storage) Save(metric metric.Metric) {
	storage.metrics[metric.GetName()] = metric.Clone()
}

func New() *Storage {
	return &Storage{metrics: make(map[string]metric.Metric)}
}
