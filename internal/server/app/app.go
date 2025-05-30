// Package app
// contains dependency injection, general goroutines and start/stop logic
package app

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/pprof"
	"github.com/m1khal3v/gometheus/internal/server/config"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/rpc"
	"github.com/m1khal3v/gometheus/internal/server/storage/factory"
	"go.uber.org/zap"
)

func Start(config *config.Config) error {
	ctx := context.Background()

	//              ||
	// increment 23 \/
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

	var privKey *rsa.PrivateKey
	if config.CryptoKey != "" {
		privKey, err = readPrivateKeyFromFile(config.CryptoKey)
		if err != nil {
			return err
		}
	}

	var subnet *net.IPNet
	if config.TrustedSubnet != "" {
		_, subnet, err = net.ParseCIDR(config.TrustedSubnet)
		if err != nil {
			return err
		}
	}

	var shutdown func(ctx context.Context) error

	if config.Protocol == "http" {
		// Настройка HTTP-сервера
		server := &http.Server{
			Addr:    config.Address,
			Handler: router.New(storage, config.Key, privKey, subnet),
		}
		shutdown = func(ctx context.Context) error {
			return server.Shutdown(ctx)
		}

		go func() {
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				errCancel(err)
			}
		}()

	} else {
		opts := []rpc.ServerOption{}
		if config.Key != "" {
			opts = append(opts, rpc.WithHMAC(config.Key, "X-Signature", sha256.New))
		}

		if privKey != nil {
			opts = append(opts, rpc.WithTLS(privKey))
		}

		if subnet != nil {
			opts = append(opts, rpc.WithSubnet("X-Real-IP", subnet))
		}

		server, err := rpc.NewGRPCServer(storage, opts...)
		if err != nil {
			errCancel(err)
		}

		shutdown = func(ctx context.Context) error {
			server.Stop()

			return nil
		}

		go func() {
			if err := server.Start(config.Address); !errors.Is(err, http.ErrServerClosed) {
				errCancel(err)
			}
		}()
	}

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

		if err := shutdown(timeoutCtx); err != nil {
			logger.Logger.Error("Failed to shutdown server", zap.Error(err))
		} else {
			logger.Logger.Info("Server was shutdown successfully")
		}

		return nil
	}
}

func readPrivateKeyFromFile(filepath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("can`t decode PEM")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
