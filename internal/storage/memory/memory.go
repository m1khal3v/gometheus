package memory

import (
	storages "github.com/m1khal3v/gometheus/internal/storage"
)

type Storage struct {
	Metrics map[string]*storages.Metric
}

func (storage *Storage) Save(metric *storages.Metric) error {
	switch metric.Type {
	case storages.MetricTypeGauge:
		storage.Metrics[metric.Name] = metric
	case storages.MetricTypeCounter:
		current, ok := storage.Metrics[metric.Name]
		// Если счетчик не существует или изменился тип - [пере]записываем
		if !ok || metric.Type != current.Type {
			storage.Metrics[metric.Name] = metric
		} else {
			current.IntValue += metric.IntValue
		}
	default:
		return storages.UnknownTypeError{
			Type: metric.Type,
		}
	}

	return nil
}

func NewStorage() *Storage {
	return &Storage{Metrics: make(map[string]*storages.Metric)}
}
