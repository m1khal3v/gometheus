package agent

import (
	"context"
	"crypto/sha256"
	"github.com/m1khal3v/gometheus/internal/agent/collector"
	"github.com/m1khal3v/gometheus/internal/agent/collector/random"
	"github.com/m1khal3v/gometheus/internal/agent/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/agent/config"
	"github.com/m1khal3v/gometheus/internal/agent/queue"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/pkg/client"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/m1khal3v/gometheus/pkg/semaphore"
	"go.uber.org/zap"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

	return []collector.Collector{
		runtimeCollector,
		randomCollector,
	}, nil
}

func Start(config *config.Config) error {
	ctx := context.Background()
	suspendCtx, _ := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	queue := queue.New(3000)
	client, err := client.New(&client.Config{
		Address: config.Address,
		Signature: &client.SignatureConfig{
			Key:    config.Key,
			Hash:   sha256.New,
			Header: "HashSHA256",
		},
	})
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

	<-suspendCtx.Done()

	logger.Logger.Info("Received suspend signal. Trying to process already collected metrics...")
	processMetrics(ctx, queue, client, semaphore, config.BatchSize)
	logger.Logger.Info("Agent successfully suspended")

	return nil
}

func collectMetricsWithInterval(ctx context.Context, queue *queue.Queue, collectors []collector.Collector, pollInterval uint32) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			collectMetrics(ctx, queue, collectors)
		}
	}
}

func collectMetrics(ctx context.Context, queue *queue.Queue, collectors []collector.Collector) {
	for _, item := range collectors {
		collector := item

		select {
		case <-ctx.Done():
			return
		default:
			go func() {
				for metric := range collector.Collect() {
					queue.Push(metric)
				}
			}()
		}
	}
}

type apiClient interface {
	SaveMetricsAsJSON(ctx context.Context, requests []request.SaveMetricRequest) ([]response.SaveMetricResponse, *response.APIError, error)
}

func processMetricsWithInterval(ctx context.Context, queue *queue.Queue, client apiClient, semaphore *semaphore.Semaphore, reportInterval uint32, batchSize uint64) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processMetrics(ctx, queue, client, semaphore, batchSize)
		}
	}
}

func processMetrics(ctx context.Context, queue *queue.Queue, client apiClient, semaphore *semaphore.Semaphore, batchSize uint64) {
	var waitGroup sync.WaitGroup

	for queue.Count() > 0 {
		select {
		case <-ctx.Done():
			return
		case <-semaphore.Acquire():
			break
		}

		metrics := queue.Pop(batchSize)
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			defer semaphore.Release()

			if err := sendMetrics(ctx, client, metrics); err != nil {
				for _, metric := range metrics {
					queue.Push(metric)
				}
			}
		}()
	}

	waitGroup.Wait()
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
			logger.Logger.Panic(err.Error())

			return err
		}

		requests = append(requests, *request)
	}

	if _, apiErr, err := client.SaveMetricsAsJSON(ctx, requests); err != nil {
		logger.Logger.Warn(err.Error())
		if apiErr != nil {
			logger.Logger.Warn(
				apiErr.Message,
				zap.Int("code", apiErr.Code),
				zap.Any("details", apiErr.Details),
			)
		}

		return err
	}

	return nil
}
