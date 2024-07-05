package router

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/api"
	_middleware "github.com/m1khal3v/gometheus/internal/server/chi/middleware"
	"github.com/m1khal3v/gometheus/internal/server/storage"
)

func New(storage storage.Storage, db *sql.DB) chi.Router {
	routes := api.New(storage, db)
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(_middleware.Decompress())
	router.Use(_middleware.Compress(5, "text/html", "application/json"))
	router.Use(_middleware.ZapLogger(logger.Logger, "http"))
	router.Get("/", routes.GetAllMetrics)
	router.Get("/ping", routes.PingDB)
	router.Route("/update", func(router chi.Router) {
		router.Post("/{type}/{name}/{value}", routes.SaveMetric)
		router.Post("/", routes.JSONSaveMetric)
	})
	router.Route("/value", func(router chi.Router) {
		router.Get("/{type}/{name}", routes.GetMetric)
		router.Post("/", routes.JSONGetMetric)
	})

	return router
}
