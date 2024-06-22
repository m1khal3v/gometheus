package server

import (
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/file_dump"
	"github.com/m1khal3v/gometheus/internal/server/storage/memory"
	"net/http"
)

func Start(endpoint, fileStoragePath string, storeInterval uint32, restore bool) {
	var storage storage.Storage = memory.New()

	if fileStoragePath != "" {
		storage = file_dump.New(storage, fileStoragePath, storeInterval, restore)
	}

	if err := http.ListenAndServe(endpoint, router.New(storage)); err != nil {
		logger.Logger.Fatal(err.Error())
	}
}
