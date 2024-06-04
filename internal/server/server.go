package server

import (
	"github.com/m1khal3v/gometheus/internal/router"
	"net/http"
)

func Start() {
	err := http.ListenAndServe(`:8080`, router.NewRouter())
	if err != nil {
		panic(err)
	}
}
