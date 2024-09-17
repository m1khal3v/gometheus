// Package app
// contains dependency injection, general goroutines and start/stop logic
package app

import (
	"context"
	"crypto/sha256"
	"os/signal"
	"syscall"

	"github.com/m1khal3v/gometheus/internal/agent/config"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/pprof"
	"github.com/m1khal3v/gometheus/pkg/client"
	"github.com/m1khal3v/gometheus/pkg/queue"
	"github.com/m1khal3v/gometheus/pkg/semaphore"
	"go.uber.org/zap"
)

func Start(config *config.Config) error {
	ctx := context.Background()
	suspendCtx, _ := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	queue := queue.New[metric.Metric](10000)

	options := make([]client.ConfigOption, 0, 1)
	if config.Key != "" {
		options = append(options, client.WithHMACSignature(config.Key, sha256.New, "HashSHA256"))
	}

	client := client.New(config.Address, options...)
	collectors, err := createCollectors()
	if err != nil {
		return err
	}

	semaphore := semaphore.New(config.RateLimit)

	go collectMetricsWithInterval(suspendCtx, queue, collectors, config.PollInterval)
	go processMetricsWithInterval(suspendCtx, queue, client, semaphore, config.ReportInterval, config.BatchSize)
	go pprof.ListenSignals(suspendCtx, config.CPUProfileFile, config.CPUProfileDuration, config.MemProfileFile)

	<-suspendCtx.Done()

	logger.Logger.Info("Received suspend signal. Trying to process already collected metrics...")
	if err := processMetrics(ctx, queue, client, semaphore, config.BatchSize); err != nil {
		logger.Logger.Error("Failed to process already collected metrics", zap.Error(err))
	}
	logger.Logger.Info("Agent successfully suspended")

	return nil
}
