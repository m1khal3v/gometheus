package memory

import (
	storages "github.com/m1khal3v/gometheus/internal/storage"
	"github.com/m1khal3v/gometheus/internal/store"
)

type Storage struct {
	metrics map[string]store.Metric
}

func (storage *Storage) Get(name string) (*store.Metric, error) {
	value, ok := storage.metrics[name]
	if !ok {
		return nil, storages.NewMetricNotFoundError(name)
	}

	return &value, nil
}

func (storage *Storage) GetAll() (map[string]store.Metric, error) {
	return storage.metrics, nil
}

func (storage *Storage) Save(metric *store.Metric) error {
	err := store.ValidateMetricType(metric.Type)
	if err != nil {
		return err
	}

	switch metric.Type {
	case store.MetricTypeGauge:
		storage.metrics[metric.Name] = *metric
	case store.MetricTypeCounter:
		current, ok := storage.metrics[metric.Name]
		// Если счетчик существует и тип не изменился складываем значения
		if ok && metric.Type == current.Type {
			metric.IntValue += current.IntValue
		}
		storage.metrics[metric.Name] = *metric
	}

	return nil
}

func NewStorage() *Storage {
	return &Storage{metrics: make(map[string]store.Metric)}
}
