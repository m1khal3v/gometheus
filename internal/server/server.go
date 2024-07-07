package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/chi/router"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/factory"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start(endpoint, fileStoragePath, databaseDriver, databaseDSN string, storeInterval uint32, restore bool) error {
	ctx := context.Background()
	storage, err := factory.New(fileStoragePath, databaseDriver, databaseDSN, storeInterval, restore)
	if err != nil {
		return err
	}

	server := &http.Server{Addr: endpoint, Handler: router.New(storage)}
	suspended := hookSuspendSignals(ctx, storage, server)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-suspended

	return nil
}

func hookSuspendSignals(ctx context.Context, storage storage.Storage, server *http.Server) <-chan struct{} {
	ctx, cancel := context.WithCancel(ctx)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		signal := <-signalChannel
		logger.Logger.Info(fmt.Sprintf("Received suspend signal: %s", signal.String()))

		if err := storage.Close(); err != nil {
			logger.Logger.Error(err.Error())
		} else {
			logger.Logger.Info("Storage was closed successfully")
		}

		if err := server.Shutdown(ctx); err != nil {
			logger.Logger.Error(err.Error())
		} else {
			logger.Logger.Info("Server shutdown successfully")
		}

		cancel()
	}()

	return ctx.Done()
}
