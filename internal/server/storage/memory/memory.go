package memory

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"sync"
)

type Storage struct {
	metrics *sync.Map
}

func New() *Storage {
	return &Storage{
		metrics: &sync.Map{},
	}
}

func (storage *Storage) Get(name string) metric.Metric {
	value, ok := storage.metrics.Load(name)
	if !ok {
		return nil
	}

	return value.(metric.Metric).Clone()
}

func (storage *Storage) GetAll() map[string]metric.Metric {
	clone := make(map[string]metric.Metric)
	storage.metrics.Range(func(key, value interface{}) bool {
		clone[key.(string)] = value.(metric.Metric).Clone()
		return true
	})

	return clone
}

func (storage *Storage) Save(metric metric.Metric) {
	storage.metrics.Store(metric.GetName(), metric.Clone())
}
