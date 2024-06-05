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

var address string

func init() {
	flag.StringVarP(&address, "address", "a", "localhost:8080", "address of gometheus server")
}

func main() {
	defer logger.Logger.Sync()

	flag.Parse()
	config := Config{
		Address: address,
	}
	err := env.Parse(&config)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	server.Start(config.Address)
}
