package main

import (
	"github.com/m1khal3v/gometheus/internal/agent"
	"github.com/m1khal3v/gometheus/internal/agent/config"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"go.uber.org/zap"
)

func main() {
	config := config.ParseConfig()
	logger.Init("agent", config.LogLevel)
	defer logger.Logger.Sync()
	defer logger.RecoverAndPanic()
	logger.Logger.Info(
		"Starting",
		zap.String("log_level", config.LogLevel),
		zap.String("address", config.Address),
		zap.Uint32("poll_interval", config.PollInterval),
		zap.Uint32("report_interval", config.ReportInterval),
		zap.Uint64("batch_size", config.BatchSize),
		zap.Uint64("rate_limit", config.RateLimit),
		zap.String("cpu_profile_file", config.CPUProfileFile),
		zap.Duration("cpu_profile_duration", config.CPUProfileDuration),
		zap.String("mem_profile_file", config.MemProfileFile),
		zap.Bool("key", config.Key != ""),
	)

	if err := agent.Start(config); err != nil {
		logger.Logger.Fatal(err.Error())
	}
}
