package server

import (
	"github.com/m1khal3v/gometheus/internal/route"
	"github.com/m1khal3v/gometheus/internal/storage/memory"
	"net/http"
)

func Start() {
	routeContainer := route.NewRouteContainer(memory.NewStorage())
	router := http.NewServeMux()
	router.HandleFunc("/update/{type}/{name}/{value}", routeContainer.SaveMetric)

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
