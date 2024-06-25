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
	if metric != nil && metric.GetType() == metricType {
		return metric
	}

	return nil
}

func (manager *Manager) GetAll() map[string]metric.Metric {
	return manager.storage.GetAll()
}

func (manager *Manager) Save(metric metric.Metric) metric.Metric {
	switch metric.GetType() {
	case gauge.Type:
		manager.storage.Save(metric)
	case counter.Type:
		manager.saveCounter(metric.(*counter.Metric))
	default:
		logger.Logger.Panic(fmt.Sprintf("Unsupported metric type %s", metric.GetType()))
	}

	return metric
}

func (manager *Manager) saveCounter(metric *counter.Metric) {
	manager.mutex.Lock(metric.GetName())
	defer manager.mutex.Unlock(metric.GetName())

	previous := manager.storage.Get(metric.GetName())
	if previous != nil && previous.GetType() == metric.GetType() {
		metric.Add(previous.(*counter.Metric).GetValue())
	}
	manager.storage.Save(metric)
}
