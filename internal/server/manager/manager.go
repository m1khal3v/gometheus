package manager

import (
	"context"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/mutex"
	"golang.org/x/exp/maps"
	"sort"
)

type ErrUnknownMetricType struct {
	MetricType string
}

func (err ErrUnknownMetricType) Error() string {
	return fmt.Sprintf("unsupported metric type: %s", err.MetricType)
}

func newErrUnknownMetricType(metricType string) error {
	return &ErrUnknownMetricType{
		MetricType: metricType,
	}
}

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

func (manager *Manager) Get(ctx context.Context, metricType, metricName string) (metric.Metric, error) {
	metric, err := manager.storage.Get(ctx, metricName)
	if err != nil {
		return nil, err
	}
	if metric == nil || metric.Type() != metricType {
		return nil, nil
	}

	return metric, nil
}

func (manager *Manager) GetAll(ctx context.Context) (<-chan metric.Metric, error) {
	return manager.storage.GetAll(ctx)
}

func (manager *Manager) Save(ctx context.Context, metric metric.Metric) (metric.Metric, error) {
	switch metric.Type() {
	case gauge.MetricType:
	case counter.MetricType:
		manager.mutex.Lock(metric.Name())
		defer manager.mutex.Unlock(metric.Name())

		if err := manager.prepareCounter(ctx, metric.(*counter.Metric), nil); err != nil {
			return nil, err
		}
	default:
		return nil, newErrUnknownMetricType(metric.Type())
	}

	if err := manager.storage.Save(ctx, metric); err != nil {
		return nil, err
	}

	return metric, nil
}

func (manager *Manager) SaveBatch(ctx context.Context, metrics []metric.Metric) ([]metric.Metric, error) {
	processed := map[string]metric.Metric{}
	// Sorting metrics to avoid deadlock
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name() < metrics[j].Name()
	})

	for _, metric := range metrics {
		previous := processed[metric.Name()]

		switch metric.Type() {
		case gauge.MetricType:
		case counter.MetricType:
			if previous == nil {
				manager.mutex.Lock(metric.Name())
				defer manager.mutex.Unlock(metric.Name())
			}

			if err := manager.prepareCounter(ctx, metric.(*counter.Metric), previous); err != nil {
				return nil, err
			}
		default:
			return nil, newErrUnknownMetricType(metric.Type())
		}

		processed[metric.Name()] = metric
	}

	metrics = maps.Values(processed)

	if err := manager.storage.SaveBatch(ctx, metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (manager *Manager) prepareCounter(ctx context.Context, metric *counter.Metric, previous metric.Metric) error {
	if previous == nil {
		var err error
		if previous, err = manager.storage.Get(ctx, metric.Name()); err != nil {
			return err
		}
	}

	if previous != nil && previous.Type() == metric.Type() {
		metric.Add(previous.(*counter.Metric).GetValue())
	}

	return nil
}

func (manager *Manager) PingStorage(ctx context.Context) error {
	return manager.storage.Ping(ctx)
}
