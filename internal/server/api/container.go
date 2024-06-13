package api

import (
	"github.com/m1khal3v/gometheus/internal/server/manager"
	"github.com/m1khal3v/gometheus/internal/server/storage"
)

type Container struct {
	manager *manager.Manager
}

func New(storage storage.Storage) *Container {
	return &Container{
		manager: manager.New(storage),
	}
}
