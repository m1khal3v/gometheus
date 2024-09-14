package app

import (
	"context"
	"time"

	"github.com/m1khal3v/gometheus/internal/agent/collector"
	"github.com/m1khal3v/gometheus/internal/agent/collector/gopsutil"
	"github.com/m1khal3v/gometheus/internal/agent/collector/random"
	"github.com/m1khal3v/gometheus/internal/agent/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/pkg/queue"
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
