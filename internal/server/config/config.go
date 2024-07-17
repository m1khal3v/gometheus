package config

import (
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   uint32 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDriver  string `env:"DATABASE_DRIVER"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVarP(&config.Address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Uint32VarP(&config.StoreInterval, "store-interval", "i", 300, "dump metrics to file interval in seconds")
	flag.StringVarP(&config.FileStoragePath, "file-storage-path", "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVarP(&config.Restore, "restore", "r", true, "restore metrics from file")
	flag.StringVar(&config.DatabaseDriver, "database-driver", "pgx", "database driver")
	flag.StringVarP(&config.DatabaseDSN, "database-dsn", "d", "", "database dsn")
	flag.Parse()
	if err := env.Parse(config); err != nil {
		panic(err)
	}

	return config
}
