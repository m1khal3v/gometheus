// Package api
// contains controllers and json decode/encode logic
package api

import (
	"github.com/m1khal3v/gometheus/internal/server/manager"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/templates"
)

type Container struct {
	manager   *manager.Manager
	templates *templates.Storage
}

func New(storage storage.Storage) *Container {
	return &Container{
		manager:   manager.New(storage),
		templates: templates.New(),
	}
}
