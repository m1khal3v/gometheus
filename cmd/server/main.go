package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/server"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

func parseConfig() Config {
	config := Config{}
	flag.StringVarP(&config.Address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Parse()
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	return config
}

func main() {
	config := parseConfig()
	if err := logger.Init("server", config.LogLevel); err != nil {
		panic(err)
	}

	defer logger.Logger.Sync()
	server.Start(config.Address)
}
