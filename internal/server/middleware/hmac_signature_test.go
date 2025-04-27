package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Тест функции HMACSignatureRespond
func TestHMACSignatureRespond(t *testing.T) {
	header := "X-Signature"
	key := "test-key"
	hashFn := sha256.New

	middleware := HMACSignatureRespond(header, hashFn, key)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"test response"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rec, req)

	result := rec.Result()
	defer result.Body.Close()

	signature := result.Header.Get(header)
	if signature == "" {
		t.Fatal("expected signature in response header")
	}

	// Проверяем, валидна ли HMAC-подпись
	expectedHash := hmacFromBody(hashFn, key, `{"message":"test response"}`)
	if signature != expectedHash {
		t.Errorf("expected signature %s, got %s", expectedHash, signature)
	}
}

// Тест функции HMACSignatureValidate
func TestHMACSignatureValidate(t *testing.T) {
	header := "X-Signature"
	key := "test-key"
	hashFn := sha256.New

	middleware := HMACSignatureValidate(header, hashFn, key)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"OK"}`))
	})

	// Формируем тело запроса
	body := `{"message":"test request"}`
	signature := hmacFromBody(hashFn, key, body)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(header, signature)
	rec := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Проверяем ошибочный запрос
	invalidReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	invalidReq.Header.Set(header, "invalid-signature")
	invalidRec := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(invalidRec, invalidReq)

	invalidRes := invalidRec.Result()
	defer invalidRes.Body.Close()

	if invalidRes.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, invalidRes.StatusCode)
	}
}

// Вспомогательная функция для вычисления HMAC
func hmacFromBody(hash func() hash.Hash, key string, body string) string {
	hmacEncoder := hmac.New(hash, []byte(key))
	hmacEncoder.Write([]byte(body))
	return hex.EncodeToString(hmacEncoder.Sum(nil))
}

// Тест на проверку ошибки чтения тела запроса
func TestHMACSignatureValidate_InvalidBody(t *testing.T) {
	header := "X-Signature"
	key := "test-key"
	hashFn := sha256.New

	middleware := HMACSignatureValidate(header, hashFn, key)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	errRead := errors.New("read error")
	badBody := &errorReader{err: errRead}
	req := httptest.NewRequest(http.MethodPost, "/", badBody)
	req.Header.Set(header, "dummy-signature")
	rec := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// Вспомогательный тип для эмуляции ошибочного тела запроса
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}
