package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/server"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func parseConfig() Config {
	config := Config{}
	flag.StringVarP(&config.Address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	return config
}

func main() {
	defer logger.Logger.Sync()
	config := parseConfig()
	server.Start(config.Address)
}
