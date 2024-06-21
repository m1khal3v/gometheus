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
	LogLevel       string `env:"LOG_LEVEL"`
}

func parseConfig() Config {
	config := Config{}
	flag.StringVarP(&config.Address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.Uint32VarP(&config.PollInterval, "poll-interval", "p", 2, "interval of collecting metrics")
	flag.Uint32VarP(&config.ReportInterval, "report-interval", "r", 10, "interval of reporting metrics")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Parse()
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	return config
}

func main() {
	config := parseConfig()
	if err := logger.Init("agent", config.LogLevel); err != nil {
		panic(err)
	}

	defer logger.Logger.Sync()
	agent.Start(config.Address, config.PollInterval, config.ReportInterval)
}
