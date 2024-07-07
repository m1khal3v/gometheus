package storage

import (
	"errors"
	"github.com/m1khal3v/gometheus/internal/common/metric"
)

var ErrStorageClosed = errors.New("storage closed")

type Storage interface {
	Save(metric metric.Metric) error
	Get(name string) (metric.Metric, error)
	GetAll() (<-chan metric.Metric, error)
	Ok() bool
	Close() error
	Reset() error
}
