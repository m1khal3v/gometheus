package memory

import (
	"context"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	store "github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/generator"
	"sync"
)

type Storage struct {
	metrics *sync.Map
	mutex   *sync.Mutex
	closed  bool
}

func New() *Storage {
	return &Storage{
		metrics: &sync.Map{},
		mutex:   &sync.Mutex{},
		closed:  false,
	}
}

func (storage *Storage) Get(ctx context.Context, name string) (metric.Metric, error) {
	value, ok := storage.metrics.Load(name)
	if !ok {
		return nil, nil
	}

	return value.(metric.Metric).Clone(), nil
}

func (storage *Storage) GetAll(ctx context.Context) (<-chan metric.Metric, error) {
	if err := storage.checkStorageClosed(); err != nil {
		return nil, err
	}

	values := generator.NewFromSyncMapOnlyValueWithContext(
		ctx,
		storage.metrics,
		func(value metric.Metric) metric.Metric {
			return value.Clone()
		},
	)

	return values, nil
}

func (storage *Storage) Save(ctx context.Context, metric metric.Metric) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	storage.metrics.Store(metric.Name(), metric.Clone())

	return nil
}

func (storage *Storage) SaveBatch(ctx context.Context, metrics []metric.Metric) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	for _, metric := range metrics {
		storage.metrics.Store(metric.Name(), metric.Clone())
	}

	return nil
}

func (storage *Storage) Ping(ctx context.Context) error {
	return storage.checkStorageClosed()
}

func (storage *Storage) Close(ctx context.Context) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if storage.closed {
		return store.ErrStorageClosed
	}

	storage.closed = true
	return nil
}

func (storage *Storage) Reset(ctx context.Context) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	storage.metrics.Range(func(key, value any) bool {
		storage.metrics.Delete(key)

		return true
	})

	return nil
}

func (storage *Storage) checkStorageClosed() error {
	if storage.closed {
		return store.ErrStorageClosed
	}

	return nil
}
