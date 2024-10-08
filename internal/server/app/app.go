// Package app
// contains dependency injection, general goroutines and start/stop logic
package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/pprof"
	"github.com/m1khal3v/gometheus/internal/server/config"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage/factory"
	"go.uber.org/zap"
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
		Handler: router.New(storage, config.Key),
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCancel(err)
		}
	}()
	go pprof.ListenSignals(suspendCtx, config.CPUProfileFile, config.CPUProfileDuration, config.MemProfileFile)

	select {
	case <-errCtx.Done():
		return context.Cause(errCtx)
	case <-suspendCtx.Done():
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		logger.Logger.Info("Received suspend signal. Trying to shutdown gracefully...")

		if err := storage.Close(timeoutCtx); err != nil {
			logger.Logger.Error("Failed to close storage", zap.Error(err))
		} else {
			logger.Logger.Info("Storage was closed successfully")
		}

		if err := server.Shutdown(timeoutCtx); err != nil {
			logger.Logger.Error("Failed to shutdown server", zap.Error(err))
		} else {
			logger.Logger.Info("Server was shutdown successfully")
		}

		return nil
	}
}
