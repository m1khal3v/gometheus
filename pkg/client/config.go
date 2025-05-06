package client

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"hash"
	"net/http"
	"net/url"
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
	baseURL   *url.URL
	signature *signatureConfig

	compress  bool
	retry     bool
	realIP    bool
	publicKey *rsa.PublicKey

	transport http.RoundTripper
}

type ConfigOption func(*config)

func newConfig(address string, options ...ConfigOption) *config {
	config := &config{
		baseURL: &url.URL{
			Scheme: "http",
			Host:   address,
		},
		compress:  true,
		retry:     true,
		transport: http.DefaultTransport,
	}

	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err == nil {
			config.baseURL = url
		}
	}

	for _, option := range options {
		option(config)
	}

	return config
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

func WithoutRealIP() ConfigOption {
	return func(config *config) {
		config.realIP = false
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

func WithAsymmetricCrypt(key []byte) ConfigOption {
	return func(config *config) {
		block, _ := pem.Decode(key)
		if block == nil {
			return
		}

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return
		}

		config.publicKey = pubKey.(*rsa.PublicKey)
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
