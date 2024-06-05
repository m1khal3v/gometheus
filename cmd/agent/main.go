package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/m1khal3v/gometheus/internal/agent"
	"github.com/m1khal3v/gometheus/internal/logger"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   uint32 `env:"POLL_INTERVAL"`
	ReportInterval uint32 `env:"REPORT_INTERVAL"`
}

var address string
var pollInterval uint32
var reportInterval uint32

func init() {
	flag.StringVarP(&address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.Uint32VarP(&pollInterval, "poll-interval", "p", 2, "interval of collecting metrics")
	flag.Uint32VarP(&reportInterval, "report-interval", "r", 10, "interval of reporting metrics")
}

func main() {
	defer logger.Logger.Sync()

	flag.Parse()
	config := Config{
		Address:        address,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}
	err := env.Parse(&config)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	agent.Start(address, pollInterval, reportInterval)
}
