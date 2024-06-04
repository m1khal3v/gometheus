package memory

import (
	storages "github.com/m1khal3v/gometheus/internal/storage"
	"github.com/m1khal3v/gometheus/internal/store"
)

type Storage struct {
	Metrics map[string]*store.Metric
}

func (storage *Storage) Get(name string) (*store.Metric, error) {
	value, ok := storage.Metrics[name]
	if !ok {
		return nil, storages.NewMetricNotFoundError(name)
	}

	return value, nil
}

func (storage *Storage) GetAll() (map[string]*store.Metric, error) {
	return storage.Metrics, nil
}

func (storage *Storage) Save(metric *store.Metric) error {
	err := store.ValidateMetricType(metric.Type)
	if err != nil {
		return err
	}

	switch metric.Type {
	case store.MetricTypeGauge:
		storage.Metrics[metric.Name] = metric
	case store.MetricTypeCounter:
		current, ok := storage.Metrics[metric.Name]
		// Если счетчик не существует или изменился тип - [пере]записываем
		if !ok || metric.Type != current.Type {
			storage.Metrics[metric.Name] = metric
		} else {
			current.IntValue += metric.IntValue
		}
	}

	return nil
}

func NewStorage() *Storage {
	return &Storage{Metrics: make(map[string]*store.Metric)}
}
