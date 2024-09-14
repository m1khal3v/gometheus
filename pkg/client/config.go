package client

import (
	"fmt"
	"hash"
	"net"
	"net/http"
)

type SignatureConfig struct {
	Key    string
	Hasher func() hash.Hash
	Header string

	SignRequest      bool
	ValidateResponse bool
}

type SignatureConfigOption func(*SignatureConfig)

type Config struct {
	Scheme    string
	Host      string
	Port      string
	Signature *SignatureConfig

	Compress bool
	Retry    bool

	address   string
	transport http.RoundTripper
}

type ConfigOption func(*Config)

func NewConfig(address string, options ...ConfigOption) *Config {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = "80"
	}

	config := &Config{
		Scheme:    "http",
		Host:      host,
		Port:      port,
		Compress:  true,
		Retry:     true,
		transport: http.DefaultTransport,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

func (config *Config) Address() string {
	if config.address == "" {
		config.address = fmt.Sprintf("%s://%s", config.Scheme, net.JoinHostPort(config.Host, config.Port))
	}

	return config.address
}

func WithScheme(scheme string) ConfigOption {
	return func(config *Config) {
		config.Scheme = scheme
	}
}

func WithPort(port uint32) ConfigOption {
	return func(config *Config) {
		config.Port = fmt.Sprintf("%d", port)
	}
}

func WithoutCompress() ConfigOption {
	return func(config *Config) {
		config.Compress = false
	}
}

func WithoutRetry() ConfigOption {
	return func(config *Config) {
		config.Retry = false
	}
}

func WithHMACSignature(key string, hasher func() hash.Hash, header string, options ...SignatureConfigOption) ConfigOption {
	return func(config *Config) {
		config.Signature = &SignatureConfig{
			Key:              key,
			Hasher:           hasher,
			Header:           header,
			SignRequest:      true,
			ValidateResponse: true,
		}

		for _, option := range options {
			option(config.Signature)
		}

		config.transport = hmacTransport
	}
}

func WithoutSignRequest() SignatureConfigOption {
	return func(config *SignatureConfig) {
		config.SignRequest = false
	}
}

func WithoutValidateResponse() SignatureConfigOption {
	return func(config *SignatureConfig) {
		config.ValidateResponse = false
	}
}
