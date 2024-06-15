package server

import (
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage/memory"
	"net/http"
)

func Start(endpoint string) {
	err := http.ListenAndServe(endpoint, router.New(memory.New()))
	if err != nil {
		logger.Logger.Panic(err.Error())
	}
}
