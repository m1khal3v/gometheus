// Package app
// contains dependency injection, general goroutines and start/stop logic
package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1khal3v/gometheus/internal/agent/collector"
	"github.com/m1khal3v/gometheus/internal/agent/collector/gopsutil"
	"github.com/m1khal3v/gometheus/internal/agent/collector/random"
	"github.com/m1khal3v/gometheus/internal/agent/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/agent/config"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/internal/common/pprof"
	"github.com/m1khal3v/gometheus/pkg/client"
	"github.com/m1khal3v/gometheus/pkg/queue"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/m1khal3v/gometheus/pkg/semaphore"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func createCollectors() ([]collector.Collector, error) {
	runtimeCollector, err := runtime.New([]string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	})
	if err != nil {
		return nil, err
	}

	randomCollector, err := random.New(0, 512)
	if err != nil {
		return nil, err
	}

	gopsUtilCollector, err := gopsutil.New(gopsutil.MetricMap{
		gopsutil.TotalMemory:    "TotalMemory",
		gopsutil.FreeMemory:     "FreeMemory",
		gopsutil.CPUUtilization: "CPUUtilization",
	})
	if err != nil {
		return nil, err
	}

	return []collector.Collector{
		runtimeCollector,
		randomCollector,
		gopsUtilCollector,
	}, nil
}

func Start(config *config.Config) error {
	ctx := context.Background()
	suspendCtx, _ := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	queue := queue.New[metric.Metric](10000)

	clientConfig := &client.Config{Address: config.Address}
	if config.Key != "" {
		clientConfig.Signature = client.NewSignatureConfig("HashSHA256", config.Key, sha256.New)
	}
	client, err := client.New(clientConfig)
	if err != nil {
		return err
	}

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

func collectMetricsWithInterval(ctx context.Context, queue *queue.Queue[metric.Metric], collectors []collector.Collector, pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := collectMetrics(ctx, queue, collectors); err != nil {
				logger.Logger.Error("Failed to collect metrics", zap.Error(err))
			}
		}
	}
}

func collectMetrics(ctx context.Context, queue *queue.Queue[metric.Metric], collectors []collector.Collector) error {
	var errGroup errgroup.Group

	for _, collector := range collectors {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
			errGroup.Go(func() error {
				collected, err := collector.Collect()
				if err != nil {
					return err
				}

				queue.PushChannel(collected)

				return nil
			})
		}
	}

	return errGroup.Wait()
}

type apiClient interface {
	SaveMetricsAsJSON(ctx context.Context, requests []request.SaveMetricRequest) ([]response.SaveMetricResponse, *response.APIError, error)
}

func processMetricsWithInterval(ctx context.Context, queue *queue.Queue[metric.Metric], client apiClient, semaphore *semaphore.Semaphore, reportInterval uint32, batchSize uint64) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := processMetrics(ctx, queue, client, semaphore, batchSize); err != nil {
				logger.Logger.Warn("Failed to process metrics", zap.Error(err))
			}
		}
	}
}

func processMetrics(ctx context.Context, queue *queue.Queue[metric.Metric], client apiClient, semaphore *semaphore.Semaphore, batchSize uint64) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var errGroup errgroup.Group

	for queue.Count() > 0 {
		if err := semaphore.Acquire(timeoutCtx); err != nil {
			return err
		}

		errGroup.Go(func() error {
			defer semaphore.Release()
			return queue.RemoveBatch(batchSize, func(items []metric.Metric) error {
				return sendMetrics(timeoutCtx, client, items)
			})
		})
	}

	return errGroup.Wait()
}

func sendMetrics(ctx context.Context, client apiClient, metrics []metric.Metric) error {
	count := len(metrics)
	if count == 0 {
		return nil
	}

	requests := make([]request.SaveMetricRequest, 0, count)
	for _, metric := range metrics {
		request, err := transformer.TransformToSaveRequest(metric)
		if err != nil {
			return err
		}

		requests = append(requests, *request)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, apiErr, err := client.SaveMetricsAsJSON(timeoutCtx, requests); err != nil {
		if apiErr != nil {
			return fmt.Errorf("code: %d. %s [%v]", apiErr.Code, apiErr.Message, apiErr.Details)
		}

		return err
	}

	return nil
}
