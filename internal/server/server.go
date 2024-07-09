package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/factory"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start(endpoint, fileStoragePath, databaseDriver, databaseDSN string, storeInterval uint32, restore bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := factory.New(
		ctx,
		fileStoragePath,
		databaseDriver,
		databaseDSN,
		storeInterval,
		restore,
	)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    endpoint,
		Handler: router.New(storage),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	hookSuspendSignals(ctx, cancel, storage, server)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-ctx.Done()

	return nil
}

func hookSuspendSignals(ctx context.Context, cancel context.CancelFunc, storage storage.Storage, server *http.Server) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		signal := <-signalChannel
		logger.Logger.Info(fmt.Sprintf("Received suspend signal: %s", signal.String()))

		if err := storage.Close(ctx); err != nil {
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
}
