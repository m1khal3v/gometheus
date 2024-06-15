package agent

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"sync"
)

type storage struct {
	mutex   sync.Mutex
	metrics []_metric.Metric
}

type filter func(metric _metric.Metric) bool

func newStorage() *storage {
	return &storage{
		metrics: make([]_metric.Metric, 0),
	}
}

func (storage *storage) appendMetrics(metrics []_metric.Metric) {
	storage.mutex.Lock()
	storage.metrics = append(storage.metrics, metrics...)
	storage.mutex.Unlock()
}

// removeMetrics applies the filter function to all metrics,
// removes from the storage metrics for which the function returned true
func (storage *storage) removeMetrics(filter filter) {
	storage.mutex.Lock()
	saved := make([]_metric.Metric, 0)
	for _, metric := range storage.metrics {
		if !filter(metric) {
			saved = append(saved, metric)
		}
	}
	storage.metrics = saved
	storage.mutex.Unlock()
}
