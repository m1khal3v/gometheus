package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/leosunmo/zapchi"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/route"
	storages "github.com/m1khal3v/gometheus/internal/storage"
)

func NewRouter(storage storages.Storage) chi.Router {
	routeContainer := route.NewRouteContainer(storage)
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	// router.Use(middleware.AllowContentType("text/plain"))
	router.Use(middleware.Compress(5))
	router.Use(zapchi.Logger(logger.Logger, "router"))
	router.Get("/", routeContainer.GetAllMetrics)
	router.Post("/update/{type}/{name}/{value}", routeContainer.SaveMetric)
	router.Get("/value/{type}/{name}", routeContainer.GetMetric)

	return router
}
