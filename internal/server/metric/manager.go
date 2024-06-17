package metric

import (
	"github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/storage"
)

type Manager struct {
	storage storage.Storage
}

func NewManager(storage storage.Storage) *Manager {
	return &Manager{
		storage: storage,
	}
}

func (manager *Manager) Get(name string) metric.Metric {
	return manager.storage.Get(name)
}

func (manager *Manager) GetAll() map[string]metric.Metric {
	return manager.storage.GetAll()
}

func (manager *Manager) Save(metric metric.Metric) {
	switch metric.GetType() {
	case gauge.Type:
		manager.storage.Save(metric)
	case counter.Type:
		manager.saveCounter(metric.(*counter.Metric))
	}
}

func (manager *Manager) saveCounter(metric *counter.Metric) {
	previous := manager.storage.Get(metric.GetName())
	if previous != nil && previous.GetType() == metric.GetType() {
		metric.Add(previous.(*counter.Metric).GetValue())
	}
	manager.storage.Save(previous)
}
