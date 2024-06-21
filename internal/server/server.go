package server

import (
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage/memory"
	"net/http"
)

func Start(endpoint string) {
	if err := http.ListenAndServe(endpoint, router.New(memory.New())); err != nil {
		logger.Logger.Panic(err.Error())
	}
}
