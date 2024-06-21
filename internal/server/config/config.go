package config

import (
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVarP(&config.Address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Parse()
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	return config
}
