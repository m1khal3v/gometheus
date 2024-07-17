package server

import (
	"context"
	"errors"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/config"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage/factory"
	"net/http"
	"os/signal"
	"syscall"
)

func Start(config *config.Config) error {
	ctx := context.Background()

	suspendCtx, suspendCancel := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer suspendCancel()

	storage, err := factory.New(
		suspendCtx,
		config.FileStoragePath,
		config.DatabaseDriver,
		config.DatabaseDSN,
		config.StoreInterval,
		config.Restore,
	)
	if err != nil {
		return err
	}

	errCtx, errCancel := context.WithCancelCause(ctx)
	defer errCancel(nil)

	server := &http.Server{
		Addr:    config.Address,
		Handler: router.New(storage),
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCancel(err)
		}
	}()

	select {
	case <-errCtx.Done():
		return errCtx.Err()
	case <-suspendCtx.Done():
		logger.Logger.Info("Received suspend signal. Trying to shutdown gracefully...")

		if err := storage.Close(ctx); err != nil {
			logger.Logger.Error(err.Error())
		} else {
			logger.Logger.Info("Storage was closed successfully")
		}

		if err := server.Shutdown(ctx); err != nil {
			logger.Logger.Error(err.Error())
		} else {
			logger.Logger.Info("Server was shutdown successfully")
		}

		return nil
	}
}
