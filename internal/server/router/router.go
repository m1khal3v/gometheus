package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/api"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	middleware2 "github.com/m1khal3v/gometheus/pkg/middleware"
)

func New(storage storage.Storage) chi.Router {
	routes := api.New(storage)
	router := chi.NewRouter()
	router.Use(middleware2.ZapLogPanic(logger.Logger, "http-panic"))
	router.Use(middleware2.ZapLogRequest(logger.Logger, "http-request"))
	router.Use(middleware.RealIP)
	router.Use(middleware2.Decompress())
	router.Use(middleware2.Compress(5, "text/html", "application/json"))
	router.Get("/", routes.GetAllMetrics)
	router.Route("/ping", func(router chi.Router) {
		router.Get("/", routes.PingStorage)
	})
	router.Route("/update", func(router chi.Router) {
		router.Post("/{type}/{name}/{value}", routes.SaveMetric)
		router.Post("/", routes.JSONSaveMetric)
	})
	router.Route("/updates", func(router chi.Router) {
		router.Post("/", routes.JSONSaveMetrics)
	})
	router.Route("/value", func(router chi.Router) {
		router.Get("/{type}/{name}", routes.GetMetric)
		router.Post("/", routes.JSONGetMetric)
	})

	return router
}
