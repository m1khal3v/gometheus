package memory

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	_storage "github.com/m1khal3v/gometheus/internal/storage"
)

type Storage struct {
	metrics map[string]_metric.Metric
}

func (storage *Storage) Get(name string) (*_metric.Metric, error) {
	value, ok := storage.metrics[name]
	if !ok {
		return nil, _storage.NewMetricNotFoundError(name)
	}

	return &value, nil
}

func (storage *Storage) GetAll() (map[string]_metric.Metric, error) {
	return storage.metrics, nil
}

func (storage *Storage) Save(metric *_metric.Metric) error {
	err := _metric.ValidateMetricType(metric.Type)
	if err != nil {
		return err
	}

	switch metric.Type {
	case _metric.TypeGauge:
		storage.metrics[metric.Name] = *metric
	case _metric.TypeCounter:
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
	return &Storage{metrics: make(map[string]_metric.Metric)}
}
