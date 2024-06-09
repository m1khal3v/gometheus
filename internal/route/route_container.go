package route

import (
	"github.com/m1khal3v/gometheus/internal/storage"
)

type Container struct {
	Storage storage.Storage
}

func New(storage storage.Storage) *Container {
	return &Container{
		Storage: storage,
	}
}
