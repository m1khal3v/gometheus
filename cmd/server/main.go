package main

import (
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server"
	"github.com/m1khal3v/gometheus/internal/server/config"
)

func main() {
	config := config.ParseConfig()
	logger.Init("server", config.LogLevel)
	defer logger.Logger.Sync()

	server.Start(
		config.Address,
		config.FileStoragePath,
		config.DatabaseDriver,
		config.DatabaseDSN,
		config.StoreInterval,
		config.Restore,
	)
}
