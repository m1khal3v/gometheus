package main

import (
	"github.com/m1khal3v/gometheus/internal/agent"
	"github.com/m1khal3v/gometheus/internal/agent/config"
	"github.com/m1khal3v/gometheus/internal/common/logger"
)

func main() {
	config := config.ParseConfig()
	if err := logger.Init("agent", config.LogLevel); err != nil {
		panic(err)
	}

	defer logger.Logger.Sync()
	agent.Start(config.Address, config.PollInterval, config.ReportInterval)
}
