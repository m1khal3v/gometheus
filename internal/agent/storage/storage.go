package storage

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/pkg/slice"
	"sync"
)

type Storage struct {
	mutex   sync.Mutex
	metrics []metric.Metric
}

type filter func(metric metric.Metric) bool
type batchFilter func(metrics []metric.Metric) bool

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

// RemoveBatch applies the batchFilter function to all metrics split into batches,
// removes metric`s batches from the Storage for which the function returned true
func (storage *Storage) RemoveBatch(filter batchFilter, batchSize uint64) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	saved := make([]metric.Metric, 0)
	for metrics := range slice.Chunk(storage.metrics, batchSize) {
		if !filter(metrics) {
			saved = append(saved, metrics...)
		}
	}
	storage.metrics = saved
}
