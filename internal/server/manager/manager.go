package manager

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/mutex"
	"github.com/m1khal3v/gometheus/internal/server/storage"
)

type Manager struct {
	mutex   *mutex.NamedMutex
	storage storage.Storage
}

func New(storage storage.Storage) *Manager {
	return &Manager{
		mutex:   mutex.NewNamedMutex(),
		storage: storage,
	}
}

func (manager *Manager) Get(metricType, metricName string) metric.Metric {
	metric := manager.storage.Get(metricName)
	if metric != nil && metric.Type() == metricType {
		return metric
	}

	return nil
}

func (manager *Manager) GetAll() map[string]metric.Metric {
	return manager.storage.GetAll()
}

func (manager *Manager) Save(metric metric.Metric) metric.Metric {
	manager.mutex.Lock(metric.Name())
	defer manager.mutex.Unlock(metric.Name())

	switch metric.Type() {
	case gauge.MetricType:
		manager.storage.Save(metric)
	case counter.MetricType:
		manager.saveCounter(metric.(*counter.Metric))
	default:
		logger.Logger.Panic(fmt.Sprintf("Unsupported metric type %s", metric.Type()))
	}

	return metric
}

func (manager *Manager) saveCounter(metric *counter.Metric) {
	previous := manager.storage.Get(metric.Name())
	if previous != nil && previous.Type() == metric.Type() {
		metric.Add(previous.(*counter.Metric).GetValue())
	}
	manager.storage.Save(metric)
}
