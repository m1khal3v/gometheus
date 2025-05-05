package client

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferReader_ReadAll(t *testing.T) {
	data := []byte("test data")
	reader := &BufferReader{Reader: bytes.NewReader(data)}

	result, err := reader.ReadAll()
	assert.Nil(t, err)
	assert.Equal(t, data, result)
}

func TestTransportHook_RoundTrip(t *testing.T) {
	hook := transportHook(func(response *http.Response) (*http.Response, error) {
		assert.NotNil(t, response)
		return response, nil
	})

	httpClient := &http.Client{Transport: hook}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := httpClient.Do(req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	resp.Body.Close()
}

func TestPreRequestHookCombine(t *testing.T) {
	hook1 := func(req *http.Request) error {
		req.Header.Set("X-Hook1", "value1")
		return nil
	}
	hook2 := func(req *http.Request) error {
		req.Header.Set("X-Hook2", "value2")
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	err := preRequestHookCombine(hook1, hook2)(nil, req)
	assert.Nil(t, err)
	assert.Equal(t, "value1", req.Header.Get("X-Hook1"))
	assert.Equal(t, "value2", req.Header.Get("X-Hook2"))
}

func TestClient_EncryptRequestBody(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	client := &HTTPClient{
		config: &config{
			publicKey: &key.PublicKey,
		},
	}

	body := bytes.NewReader([]byte("secret data"))
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", body)

	err := client.encryptRequestBody(req)
	assert.Nil(t, err)
	assert.NotNil(t, req.Body)
	assert.Equal(t, "RSA-PKCS1v15", req.Header.Get("Content-Encryption"))
}

func TestClient_CompressRequestBody(t *testing.T) {
	client := &HTTPClient{
		gzipPool: &sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(io.Discard)
			},
		},
	}

	body := bytes.NewReader([]byte("compress me"))
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", body)

	err := client.compressRequestBody(req)
	assert.Nil(t, err)
	assert.NotNil(t, req.Body)
	assert.Equal(t, "gzip", req.Header.Get("Content-Encoding"))
}

func TestClient_AddHMACSignature(t *testing.T) {
	client := &HTTPClient{
		config: &config{
			signature: &signatureConfig{header: "X-Signature"},
		},
		hmacPool: &sync.Pool{
			New: func() interface{} {
				return sha256.New()
			},
		},
	}

	body := bytes.NewReader([]byte("sign this"))
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", body)

	err := client.addHMACSignature(req)
	assert.Nil(t, err)
	assert.NotNil(t, req.Header.Get("X-Signature"))
}
