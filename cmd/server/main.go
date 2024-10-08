package main

import (
	"github.com/m1khal3v/gometheus/internal/common/buildlog"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/server/app"
	"github.com/m1khal3v/gometheus/internal/server/config"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildlog.Print(buildVersion, buildDate, buildCommit)
	config := config.ParseConfig()
	logger.Init("server", config.LogLevel)
	defer logger.Logger.Sync()
	defer logger.RecoverAndPanic()
	logger.Logger.Info(
		"Starting",
		zap.String("log_level", config.LogLevel),
		zap.String("address", config.Address),
		zap.String("file_storage_path", config.FileStoragePath),
		zap.String("database_driver", config.DatabaseDriver),
		zap.String("database_dsn", config.DatabaseDSN),
		zap.Uint32("store_interval", config.StoreInterval),
		zap.Bool("restore", config.Restore),
		zap.Bool("key", config.Key != ""),
	)

	if err := app.Start(config); err != nil {
		logger.Logger.Fatal(err.Error())
	}
}
