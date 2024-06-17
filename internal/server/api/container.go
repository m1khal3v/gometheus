package api

import (
	"github.com/m1khal3v/gometheus/internal/server/metric"
	"github.com/m1khal3v/gometheus/internal/server/storage"
)

type Container struct {
	manager *metric.Manager
}

func New(storage storage.Storage) *Container {
	return &Container{
		manager: metric.NewManager(storage),
	}
}
