package api

import (
	"github.com/m1khal3v/gometheus/internal/storage"
)

type Container struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Container {
	return &Container{
		storage: storage,
	}
}
