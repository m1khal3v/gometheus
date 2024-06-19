package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/server/api"
	_middleware "github.com/m1khal3v/gometheus/internal/server/chi/middleware"
	"github.com/m1khal3v/gometheus/internal/server/storage"
)

func New(storage storage.Storage) chi.Router {
	routes := api.New(storage)
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.Compress(5))
	router.Use(_middleware.ZapLogger(logger.Logger, "http"))
	router.Get("/", routes.GetAllMetrics)
	router.Post("/update/{type}/{name}/{value}", routes.SaveMetric)
	router.Get("/value/{type}/{name}", routes.GetMetric)

	return router
}
