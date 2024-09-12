package pprof

import (
	"context"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/pkg/pprof"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ListenSignals(ctx context.Context, cpuFilename string, cpuDuration time.Duration, memFilename string) {
	channel := make(chan os.Signal, 1)
	defer close(channel)
	signal.Notify(channel, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		select {
		case signal := <-channel:
			switch signal {
			case syscall.SIGUSR1:
				go func() {
					logger.Logger.Info(
						"SIGUSR1 received. starting CPU profile capture...",
						zap.String("filename", cpuFilename),
						zap.Duration("duration", cpuDuration),
					)
					if err := pprof.CPUCapture(ctx, cpuFilename, cpuDuration); err != nil {
						logger.Logger.Error("cpu profile capture failed", zap.Error(err))
						return
					}
					logger.Logger.Info("cpu profile capture finished")
				}()
			case syscall.SIGUSR2:
				go func() {
					logger.Logger.Info("SIGUSR2 received. starting memory profile capture...", zap.String("filename", memFilename))
					if err := pprof.Capture(pprof.Heap, memFilename); err != nil {
						logger.Logger.Error("mem capture failed", zap.Error(err))
						return
					}
					logger.Logger.Info("memory profile capture finished")
				}()
			}
		case <-ctx.Done():
			return
		}
	}
}
