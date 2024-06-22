package main

import (
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server"
	"github.com/m1khal3v/gometheus/internal/server/config"
)

func main() {
	config := config.ParseConfig()
	if err := logger.Init("server", config.LogLevel); err != nil {
		panic(err)
	}

	defer logger.Logger.Sync()
	server.Start(config.Address, config.FileStoragePath, config.StoreInterval, config.Restore)
}
