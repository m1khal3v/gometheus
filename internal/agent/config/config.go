package config

import (
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   uint32 `env:"POLL_INTERVAL"`
	ReportInterval uint32 `env:"REPORT_INTERVAL"`
	LogLevel       string `env:"LOG_LEVEL"`
	BatchSize      uint64 `env:"BATCH_SIZE"`
	Key            string `env:"KEY"`
	RateLimit      uint64 `env:"RATE_LIMIT"`
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVarP(&config.Address, "address", "a", "localhost:8080", "address of gometheus server")
	flag.Uint32VarP(&config.PollInterval, "poll-interval", "p", 2, "interval of collecting metrics")
	flag.Uint32VarP(&config.ReportInterval, "report-interval", "r", 10, "interval of reporting metrics")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Uint64VarP(&config.BatchSize, "batch-size", "b", 200, "number of metrics sent within one request")
	flag.StringVarP(&config.Key, "key", "k", "", "secret key")
	flag.Uint64VarP(&config.RateLimit, "rate-limit", "m", 10, "maximum number of concurrently executing requests")
	flag.Parse()
	if err := env.Parse(config); err != nil {
		panic(err)
	}

	return config
}
