package main

import (
	"github.com/m1khal3v/gometheus/internal/agent"
	"github.com/m1khal3v/gometheus/internal/agent/config"
	"github.com/m1khal3v/gometheus/internal/common/logger"
)

func main() {
	config := config.ParseConfig()
	logger.Init("agent", config.LogLevel)
	defer logger.Logger.Sync()
	defer logger.RecoverAndPanic()

	if err := agent.Start(
		config.Address,
		config.PollInterval,
		config.ReportInterval,
	); err != nil {
		logger.Logger.Fatal(err.Error())
	}
}
