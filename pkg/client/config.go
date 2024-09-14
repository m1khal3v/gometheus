package client

import (
	"fmt"
	"hash"
	"net"
	"net/http"
	"strings"
)

type signatureConfig struct {
	key    string
	hasher func() hash.Hash
	header string

	signRequest      bool
	validateResponse bool
}

type SignatureConfigOption func(*signatureConfig)

type config struct {
	scheme    string
	host      string
	port      string
	signature *signatureConfig

	compress bool
	retry    bool

	address   string
	transport http.RoundTripper
}

type ConfigOption func(*config)

func newConfig(address string, options ...ConfigOption) *config {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = "80"
	}

	config := &config{
		scheme:    "http",
		host:      host,
		port:      port,
		compress:  true,
		retry:     true,
		transport: http.DefaultTransport,
	}

	for _, option := range options {
		option(config)
	}

	config.address = fmt.Sprintf("%s://%s", config.scheme, strings.TrimRight(net.JoinHostPort(config.host, config.port), ":"))

	return config
}

func WithScheme(scheme string) ConfigOption {
	return func(config *config) {
		config.scheme = scheme
	}
}

func WithPort(port uint32) ConfigOption {
	return func(config *config) {
		config.port = fmt.Sprintf("%d", port)
	}
}

func WithoutCompress() ConfigOption {
	return func(config *config) {
		config.compress = false
	}
}

func WithoutRetry() ConfigOption {
	return func(config *config) {
		config.retry = false
	}
}

func WithHMACSignature(key string, hasher func() hash.Hash, header string, options ...SignatureConfigOption) ConfigOption {
	return func(config *config) {
		config.signature = &signatureConfig{
			key:              key,
			hasher:           hasher,
			header:           header,
			signRequest:      true,
			validateResponse: true,
		}

		for _, option := range options {
			option(config.signature)
		}

		config.transport = hmacTransport
	}
}

func withTransport(transport http.RoundTripper) ConfigOption {
	return func(config *config) {
		config.transport = transport
	}
}

func WithoutSignRequest() SignatureConfigOption {
	return func(config *signatureConfig) {
		config.signRequest = false
	}
}

func WithoutValidateResponse() SignatureConfigOption {
	return func(config *signatureConfig) {
		config.validateResponse = false
	}
}
