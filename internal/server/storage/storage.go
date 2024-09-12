package storage

import (
	"context"
	"errors"

	"github.com/m1khal3v/gometheus/internal/common/metric"
)

var ErrStorageClosed = errors.New("storage closed")

type Storage interface {
	Save(ctx context.Context, metric metric.Metric) error
	SaveBatch(ctx context.Context, metrics []metric.Metric) error
	Get(ctx context.Context, name string) (metric.Metric, error)
	GetAll(ctx context.Context) (<-chan metric.Metric, error)
	Ping(ctx context.Context) error
	Reset(ctx context.Context) error
	Close(ctx context.Context) error
}
