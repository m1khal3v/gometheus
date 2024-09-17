// Package storage
// contains interface and its implementations
package storage

import (
	"context"
	"errors"

	"github.com/m1khal3v/gometheus/internal/common/metric"
)

var ErrStorageClosed = errors.New("storage closed")

type Storage interface {
	Save(ctx context.Context, metric metric.Metric) error         // Save one metric to Storage
	SaveBatch(ctx context.Context, metrics []metric.Metric) error // SaveBatch of metric to Storage
	Get(ctx context.Context, name string) (metric.Metric, error)  // Get metric from Storage
	GetAll(ctx context.Context) (<-chan metric.Metric, error)     // GetAll metrics from Storage
	Ping(ctx context.Context) error                               // Ping Storage connection
	Reset(ctx context.Context) error                              // Reset Storage (delete all metrics)
	Close(ctx context.Context) error                              // Close Storage (graceful shutdown)
}
