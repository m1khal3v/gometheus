// Package config
// gometheus agent configuration
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type jsonConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
}

type Config struct {
	Address            string        `env:"ADDRESS"`
	PollInterval       uint32        `env:"POLL_INTERVAL"`
	ReportInterval     uint32        `env:"REPORT_INTERVAL"`
	LogLevel           string        `env:"LOG_LEVEL"`
	BatchSize          uint64        `env:"BATCH_SIZE"`
	Key                string        `env:"KEY"`
	RateLimit          uint64        `env:"RATE_LIMIT"`
	CPUProfileFile     string        `env:"CPU_PROFILE_FILE"`
	CPUProfileDuration time.Duration `env:"CPU_PROFILE_DURATION"`
	MemProfileFile     string        `env:"MEM_PROFILE_FILE"`
	CryptoKey          string        `env:"CRYPTO_KEY"`
	Protocol           string        `env:"PROTOCOL"`
}

func ParseConfig() *Config {
	config := &Config{}
	jsonCfg, err := parseJSONConfig()
	if err != nil {
		panic(err)
	}

	defaultAddress := "localhost:8080"
	if jsonCfg != nil && jsonCfg.Address != nil {
		defaultAddress = *jsonCfg.Address
	}
	flag.StringVarP(&config.Address, "address", "a", defaultAddress, "address of gometheus server")

	defaultReportInterval := uint32(10)
	if jsonCfg != nil && jsonCfg.ReportInterval != nil {
		duration, err := time.ParseDuration(*jsonCfg.ReportInterval)
		if err != nil {
			panic(err)
		}
		defaultReportInterval = uint32(duration.Seconds())
	}
	flag.Uint32VarP(&config.ReportInterval, "report-interval", "r", defaultReportInterval, "interval of reporting metrics")

	defaultPollInterval := uint32(2)
	if jsonCfg != nil && jsonCfg.PollInterval != nil {
		duration, err := time.ParseDuration(*jsonCfg.PollInterval)
		if err != nil {
			panic(err)
		}
		defaultPollInterval = uint32(duration.Seconds())
	}
	flag.Uint32VarP(&config.PollInterval, "poll-interval", "p", defaultPollInterval, "interval of collecting metrics")

	defaultCryptoKey := ""
	if jsonCfg != nil && jsonCfg.CryptoKey != nil {
		defaultCryptoKey = *jsonCfg.CryptoKey
	}
	flag.StringVar(&config.CryptoKey, "crypto-key", defaultCryptoKey, "path to private key")

	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Uint64VarP(&config.BatchSize, "batch-size", "b", 200, "number of metrics sent within one request")
	flag.StringVarP(&config.Key, "key", "k", "", "secret key")
	flag.Uint64VarP(&config.RateLimit, "rate-limit", "m", 10, "maximum number of concurrently executing requests")
	flag.StringVar(&config.CPUProfileFile, "cpu-profile-file", "cpu.pprof", "path to save CPU profile")
	flag.DurationVar(&config.CPUProfileDuration, "cpu-profile-duration", time.Second*30, "duration to save CPU profile")
	flag.StringVar(&config.MemProfileFile, "mem-profile-file", "mem.pprof", "path to save memory profile")
	flag.StringVar(&config.Protocol, "protocol", "http", "http/grpc")
	flag.Parse()

	if err := env.Parse(config); err != nil {
		panic(err)
	}

	if config.Protocol != "http" && config.Protocol != "grpc" {
		panic("invalid protocol")
	}

	return config
}

func parseJSONConfig() (*jsonConfig, error) {
	var configFile string
	args := os.Args[1:]

	// Ручная обработка аргументов для извлечения -c/--config
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-c" || arg == "--config" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("config flag requires a value")
			}
			configFile = args[i+1]
			// Удаляем обработанные аргументы из os.Args
			os.Args = append(os.Args[:1+i], os.Args[i+2:]...)
			break
		}

		if strings.HasPrefix(arg, "--config=") {
			configFile = strings.TrimPrefix(arg, "--config=")
			os.Args = append(os.Args[:1+i], os.Args[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "-c=") {
			configFile = strings.TrimPrefix(arg, "-c=")
			os.Args = append(os.Args[:1+i], os.Args[i+1:]...)
			break
		}
	}

	// Проверяем переменную окружения CONFIG
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configFile = envConfig
	}

	if configFile == "" {
		return nil, nil
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &jsonConfig{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
