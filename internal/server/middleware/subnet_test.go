package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockWriter для проверки вызова функций api.WriteJSONErrorResponse
type MockWriter struct {
	http.ResponseWriter
	status       int
	responseBody string
}

func (m *MockWriter) WriteHeader(code int) {
	m.status = code
}

func (m *MockWriter) Write(b []byte) (int, error) {
	m.responseBody = string(b)
	return len(b), nil
}

func TestSubnetValidate(t *testing.T) {
	// Создаем моковый сабнет
	_, subnet, _ := net.ParseCIDR("192.168.0.0/24")
	middleware := SubnetValidate("X-Real-IP", subnet)

	// Моковый обработчик
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name            string
		ipHeader        string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "Valid IP in subnet",
			ipHeader:        "192.168.0.10",
			expectedStatus:  http.StatusOK,
			expectedMessage: "success",
		},
		{
			name:            "Missing IP header",
			ipHeader:        "",
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "IP header is missing",
		},
		{
			name:            "Invalid IP format",
			ipHeader:        "invalid-ip",
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "Invalid IP format",
		},
		{
			name:            "IP not in allowed subnet",
			ipHeader:        "10.0.0.1",
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "IP not in allowed subnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем моковый запрос и запись
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.ipHeader != "" {
				req.Header.Set("X-Real-IP", tt.ipHeader)
			}
			recorder := httptest.NewRecorder()

			// Оборачиваем моковый обработчик мидлварой
			handler := middleware(mockHandler)
			handler.ServeHTTP(recorder, req)

			// Проверяем статус и сообщение
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if recorder.Code != http.StatusOK {
				assert.Contains(t, recorder.Body.String(), tt.expectedMessage)
			} else {
				assert.Equal(t, tt.expectedMessage, recorder.Body.String())
			}
		})
	}
}
