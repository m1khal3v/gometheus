package app

import (
	"context"
	"fmt"
	"time"

	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/pkg/client"
	"github.com/m1khal3v/gometheus/pkg/queue"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/semaphore"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func processMetricsWithInterval(ctx context.Context, queue *queue.Queue[metric.Metric], client client.Client, semaphore *semaphore.Semaphore, reportInterval uint32, batchSize uint64) {
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

func processMetrics(ctx context.Context, queue *queue.Queue[metric.Metric], client client.Client, semaphore *semaphore.Semaphore, batchSize uint64) error {
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

func sendMetrics(ctx context.Context, client client.Client, metrics []metric.Metric) error {
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
	if _, apiErr, err := client.SaveMetrics(timeoutCtx, requests); err != nil {
		if apiErr != nil {
			return fmt.Errorf("code: %d. %s [%v]", apiErr.Code, apiErr.Message, apiErr.Details)
		}

		return err
	}

	return nil
}
