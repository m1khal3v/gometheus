package server

import (
	"github.com/m1khal3v/gometheus/internal/router"
	"github.com/m1khal3v/gometheus/internal/storage/memory"
	"net/http"
)

func Start(endpoint string) {
	err := http.ListenAndServe(endpoint, router.NewRouter(memory.NewStorage()))
	if err != nil {
		panic(err)
	}
}
