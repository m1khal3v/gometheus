package api

import (
	"database/sql"
	"github.com/m1khal3v/gometheus/internal/server/manager"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/templates"
)

type Container struct {
	db        *sql.DB
	manager   *manager.Manager
	templates *templates.Storage
}

func New(storage storage.Storage, db *sql.DB) *Container {
	return &Container{
		db:        db,
		manager:   manager.New(storage),
		templates: templates.New(),
	}
}
