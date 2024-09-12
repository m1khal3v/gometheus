// Package config
// gometheus server configuration
package config

import (
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address            string        `env:"ADDRESS"`
	LogLevel           string        `env:"LOG_LEVEL"`
	StoreInterval      uint32        `env:"STORE_INTERVAL"`
	FileStoragePath    string        `env:"FILE_STORAGE_PATH"`
	Restore            bool          `env:"RESTORE"`
	DatabaseDriver     string        `env:"DATABASE_DRIVER"`
	DatabaseDSN        string        `env:"DATABASE_DSN"`
	Key                string        `env:"KEY"`
	CPUProfileFile     string        `env:"CPU_PROFILE_FILE"`
	CPUProfileDuration time.Duration `env:"CPU_PROFILE_DURATION"`
	MemProfileFile     string        `env:"MEM_PROFILE_FILE"`
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
	flag.StringVarP(&config.Key, "key", "k", "", "secret key")
	flag.StringVar(&config.CPUProfileFile, "cpu-profile-file", "cpu.pprof", "path to save CPU profile")
	flag.DurationVar(&config.CPUProfileDuration, "cpu-profile-duration", time.Second*30, "duration to save CPU profile")
	flag.StringVar(&config.MemProfileFile, "mem-profile-file", "mem.pprof", "path to save memory profile")
	flag.Parse()
	if err := env.Parse(config); err != nil {
		panic(err)
	}

	return config
}
