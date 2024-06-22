package server

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/dump"
	"github.com/m1khal3v/gometheus/internal/server/storage/memory"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start(endpoint, fileStoragePath string, storeInterval uint32, restore bool) {
	var storage storage.Storage = memory.New()

	if fileStoragePath != "" {
		storage = dump.New(storage, fileStoragePath, storeInterval, restore)

		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			signal := <-signalChannel
			logger.Logger.Info(fmt.Sprintf("Received suspend signal: %s", signal.String()))
			storage.(*dump.Storage).Dump()
			os.Exit(0)
		}()
	}

	if err := http.ListenAndServe(endpoint, router.New(storage)); err != nil {
		logger.Logger.Fatal(err.Error())
	}
}
