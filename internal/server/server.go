package server

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/chi/router"
	"github.com/m1khal3v/gometheus/internal/server/db"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/dump"
	"github.com/m1khal3v/gometheus/internal/server/storage/memory"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start(endpoint, fileStoragePath, databaseDriver, databaseDSN string, storeInterval uint32, restore bool) {
	var storage storage.Storage = memory.New()
	var database *sql.DB = nil

	if databaseDSN != "" && databaseDriver != "" {
		var err error = nil
		database, err = db.New(databaseDriver, databaseDSN)

		if err != nil {
			logger.Logger.Panic(err.Error())
		}
	}

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

	if err := http.ListenAndServe(endpoint, router.New(storage, database)); !errors.Is(err, http.ErrServerClosed) {
		logger.Logger.Panic(err.Error())
	}
}
