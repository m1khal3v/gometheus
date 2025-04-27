package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"hash"
	"hash/fnv"
	"net/http"
	"testing"
)

func TestNewConfig_DefaultValues(t *testing.T) {
	address := "example.com"
	conf := newConfig(address)

	if conf.baseURL.Host != address {
		t.Errorf("Expected baseURL host to be %s, got %s", address, conf.baseURL.Host)
	}
	if !conf.compress {
		t.Error("Expected compress to be true by default")
	}
	if !conf.retry {
		t.Error("Expected retry to be true by default")
	}
	if conf.transport != http.DefaultTransport {
		t.Error("Expected default transport to be http.DefaultTransport")
	}
}

func TestNewConfig_WithCustomURL(t *testing.T) {
	address := "https://example.com"
	conf := newConfig(address)

	if conf.baseURL.String() != address {
		t.Errorf("Expected baseURL to be %s, got %s", address, conf.baseURL.String())
	}
}

func TestWithoutCompress(t *testing.T) {
	address := "example.com"
	conf := newConfig(address, WithoutCompress())

	if conf.compress {
		t.Error("Expected compress to be false")
	}
}

func TestWithoutRetry(t *testing.T) {
	address := "example.com"
	conf := newConfig(address, WithoutRetry())

	if conf.retry {
		t.Error("Expected retry to be false")
	}
}

func TestWithHMACSignature(t *testing.T) {
	address := "example.com"
	conf := newConfig(address, WithHMACSignature("key", func() hash.Hash {
		return fnv.New64()
	}, "Authorization"))

	if conf.signature == nil {
		t.Fatal("Expected signature configuration to be set")
	}
	if conf.signature.key != "key" {
		t.Errorf("Expected signature key to be 'key', got %s", conf.signature.key)
	}
	if conf.signature.header != "Authorization" {
		t.Errorf("Expected signature header to be 'Authorization', got %s", conf.signature.header)
	}
	if conf.signature.signRequest != true {
		t.Error("Expected signRequest to be true")
	}
	if conf.signature.validateResponse != true {
		t.Error("Expected validateResponse to be true")
	}
}

func TestWithAsymmetricCrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyBytes})
	conf := newConfig("example.com", WithAsymmetricCrypt(pemKey))

	if conf.publicKey == nil {
		t.Fatal("Expected publicKey to be set")
	}
	if conf.publicKey.E != privateKey.PublicKey.E || conf.publicKey.N.Cmp(privateKey.PublicKey.N) != 0 {
		t.Error("Public key does not match the provided one")
	}
}

func TestWithoutSignRequest(t *testing.T) {
	address := "example.com"
	conf := newConfig(address, WithHMACSignature("key", func() hash.Hash {
		return fnv.New64()
	}, "Authorization", WithoutSignRequest()))

	if conf.signature == nil {
		t.Fatal("Expected signature configuration to be set")
	}
	if conf.signature.signRequest != false {
		t.Error("Expected signRequest to be false")
	}
}

func TestWithoutValidateResponse(t *testing.T) {
	address := "example.com"
	conf := newConfig(address, WithHMACSignature("key", func() hash.Hash {
		return fnv.New64()
	}, "Authorization", WithoutValidateResponse()))

	if conf.signature == nil {
		t.Fatal("Expected signature configuration to be set")
	}
	if conf.signature.validateResponse != false {
		t.Error("Expected validateResponse to be false")
	}
}
