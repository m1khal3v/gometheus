package storage

import (
	"github.com/m1khal3v/gometheus/internal/metric"
	"sync"
)

type Storage struct {
	mutex   sync.Mutex
	metrics []metric.Metric
}

type filter func(metric metric.Metric) bool

func New() *Storage {
	return &Storage{
		metrics: make([]metric.Metric, 0),
	}
}

func (storage *Storage) Append(metrics []metric.Metric) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	for _, metric := range metrics {
		storage.metrics = append(storage.metrics, metric.Clone())
	}
}

// Remove applies the filter function to all metrics,
// removes metrics from the Storage for which the function returned true
func (storage *Storage) Remove(filter filter) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	saved := make([]metric.Metric, 0)
	for _, metric := range storage.metrics {
		if !filter(metric) {
			saved = append(saved, metric)
		}
	}
	storage.metrics = saved
}
