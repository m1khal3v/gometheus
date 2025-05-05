// Package config
// gometheus server configuration
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
	Address         *string `json:"address"`
	StoreInterval   *string `json:"store_interval"`
	FileStoragePath *string `json:"store_file"`
	Restore         *bool   `json:"restore"`
	DatabaseDSN     *string `json:"database_dsn"`
	CryptoKey       *string `json:"crypto_key"`
	TrustedSubnet   *string `json:"trusted_subnet"`
}

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
	CryptoKey          string        `env:"CRYPTO_KEY"`
	TrustedSubnet      string        `env:"TRUSTED_SUBNET"`
	Protocol           string        `env:"Protocol"` // Добавлено поле Platform
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

	defaultStoreInterval := uint32(300)
	if jsonCfg != nil && jsonCfg.StoreInterval != nil {
		duration, err := time.ParseDuration(*jsonCfg.StoreInterval)
		if err != nil {
			panic(err)
		}

		defaultStoreInterval = uint32(duration.Seconds())
	}
	flag.Uint32VarP(&config.StoreInterval, "store-interval", "i", defaultStoreInterval, "dump metrics to file interval in seconds")

	defaultFileStoragePath := "/tmp/metrics-db.json"
	if jsonCfg != nil && jsonCfg.FileStoragePath != nil {
		defaultFileStoragePath = *jsonCfg.FileStoragePath
	}
	flag.StringVarP(&config.FileStoragePath, "file-storage-path", "f", defaultFileStoragePath, "file storage path")

	defaultRestore := true
	if jsonCfg != nil && jsonCfg.Restore != nil {
		defaultRestore = *jsonCfg.Restore
	}
	flag.BoolVarP(&config.Restore, "restore", "r", defaultRestore, "restore metrics from file")

	defaultDatabaseDSN := ""
	if jsonCfg != nil && jsonCfg.DatabaseDSN != nil {
		defaultDatabaseDSN = *jsonCfg.DatabaseDSN
	}
	flag.StringVarP(&config.DatabaseDSN, "database-dsn", "d", defaultDatabaseDSN, "database dsn")

	defaultCryptoKey := ""
	if jsonCfg != nil && jsonCfg.CryptoKey != nil {
		defaultCryptoKey = *jsonCfg.CryptoKey
	}
	flag.StringVar(&config.CryptoKey, "crypto-key", defaultCryptoKey, "path to private key")

	defaultTrustedSubnet := ""
	if jsonCfg != nil && jsonCfg.TrustedSubnet != nil {
		defaultTrustedSubnet = *jsonCfg.TrustedSubnet
	}
	flag.StringVarP(&config.TrustedSubnet, "trusted-subnet", "t", defaultTrustedSubnet, "CIDR")

	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.StringVar(&config.DatabaseDriver, "database-driver", "pgx", "database driver")
	flag.StringVarP(&config.Key, "key", "k", "", "secret key")
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
