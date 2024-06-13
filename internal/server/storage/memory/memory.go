package memory

import (
	"github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/server/mutex"
)

type Storage struct {
	mutex   *mutex.NamedMutex
	metrics map[string]metric.Metric
}

func (storage *Storage) Get(name string) metric.Metric {
	storage.mutex.Lock(name)
	defer storage.mutex.Unlock(name)

	metric, ok := storage.metrics[name]
	if !ok {
		return nil
	}

	return metric.Clone()
}

func (storage *Storage) GetAll() map[string]metric.Metric {
	storage.mutex.LockAll()
	defer storage.mutex.UnlockAll()

	clone := make(map[string]metric.Metric)
	for name, metric := range storage.metrics {
		clone[name] = metric.Clone()
	}

	return clone
}

func (storage *Storage) Save(metric metric.Metric) {
	storage.mutex.Lock(metric.GetName())
	defer storage.mutex.Unlock(metric.GetName())

	storage.metrics[metric.GetName()] = metric.Clone()
}

func New() *Storage {
	return &Storage{
		mutex:   mutex.NewNamedMutex(),
		metrics: make(map[string]metric.Metric),
	}
}
