package server

import (
	"github.com/m1khal3v/gometheus/internal/router"
	"github.com/m1khal3v/gometheus/internal/storage/memory"
	"net/http"
)

func Start() {
	err := http.ListenAndServe(`:8080`, router.NewRouter(memory.NewStorage()))
	if err != nil {
		panic(err)
	}
}
