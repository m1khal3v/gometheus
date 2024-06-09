package memory

import (
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/metric/counter"
	"github.com/m1khal3v/gometheus/internal/metric/gauge"
	_storage "github.com/m1khal3v/gometheus/internal/storage"
)

type Storage struct {
	metrics map[string]_metric.Metric
}

func (storage *Storage) Get(name string) (_metric.Metric, error) {
	value, ok := storage.metrics[name]
	if !ok {
		return nil, _storage.NewMetricNotFoundError(name)
	}

	return value, nil
}

func (storage *Storage) GetAll() (map[string]_metric.Metric, error) {
	return storage.metrics, nil
}

func (storage *Storage) Save(metric _metric.Metric) error {
	switch metric := metric.(type) {
	case *gauge.Metric:
		storage.metrics[metric.GetName()] = metric
	case *counter.Metric:
		current, ok := storage.metrics[metric.GetName()]
		// if the counter exists and the type has not changed, add one to the other
		if ok && metric.GetType() == current.GetType() {
			_, err := metric.Add(current.(*counter.Metric))
			if err != nil {
				return err
			}
		}
		storage.metrics[metric.GetName()] = metric
	}

	return nil
}

func New() *Storage {
	return &Storage{metrics: make(map[string]_metric.Metric)}
}
