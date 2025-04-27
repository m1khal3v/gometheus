package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecryptMiddleware(t *testing.T) {
	// Генерация RSA-ключей для тестирования
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privKey.PublicKey

	middleware := Decrypt(privKey)

	t.Run("Valid encrypted request", func(t *testing.T) {
		originalData := "test payload"
		encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(originalData))
		if err != nil {
			t.Fatalf("Failed to encrypt data: %v", err)
		}

		encodedData := base64.StdEncoding.EncodeToString(encryptedData)

		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(encodedData)))
		req.Header.Set("Content-Encryption", "RSA-PKCS1v15")

		responseRecorder := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, originalData, string(body), "The decrypted body must match the original data")
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(responseRecorder, req)

		assert.Equal(t, http.StatusOK, responseRecorder.Code, "Expected status OK")
	})

	t.Run("Request without Content-Encryption header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test payload"))
		responseRecorder := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, "test payload", string(body), "Body should pass through unchanged")
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(responseRecorder, req)

		assert.Equal(t, http.StatusOK, responseRecorder.Code, "Expected status OK")
	})

	t.Run("Invalid base64 payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid base64"))
		req.Header.Set("Content-Encryption", "RSA-PKCS1v15")

		responseRecorder := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Handler should not be called for invalid base64")
		}))

		handler.ServeHTTP(responseRecorder, req)

		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code, "Expected status Bad Request")
		assert.Contains(t, responseRecorder.Body.String(), "Can`t base64 decode request")
	})

	t.Run("Invalid RSA decryption", func(t *testing.T) {
		payload := []byte("invalid data that cannot be decrypted")
		encodedData := base64.StdEncoding.EncodeToString(payload)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(encodedData))
		req.Header.Set("Content-Encryption", "RSA-PKCS1v15")

		responseRecorder := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Handler should not be called for invalid decryption")
		}))

		handler.ServeHTTP(responseRecorder, req)

		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code, "Expected status Bad Request")
		assert.Contains(t, responseRecorder.Body.String(), "Can`t decrypt request")
	})
}
