package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestZapLogger(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		path   string
		method string
		status int
	}{
		{
			name:   "ok",
			body:   "Hello, World!",
			path:   "/ok",
			method: http.MethodGet,
			status: http.StatusOK,
		},
		{
			name:   "not found",
			body:   "Not found!",
			path:   "/404",
			method: http.MethodPost,
			status: http.StatusNotFound,
		},
		{
			name:   "bad request",
			body:   "Bad request!",
			path:   "/400",
			method: http.MethodPut,
			status: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			core, logs := observer.New(zapcore.InfoLevel)
			logger := zap.New(core)
			router.Use(ZapLogger(logger, tt.name))
			router.MethodFunc(tt.method, tt.path, func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(tt.status)
				writer.Write([]byte(tt.body))
			})
			httpServer := httptest.NewServer(router)
			defer httpServer.Close()

			request, err := http.NewRequest(tt.method, httpServer.URL+tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()

			allLogs := logs.All()
			assert.Len(t, allLogs, 1)
			log := allLogs[0]

			assert.Equal(t, "Request processed", log.Message)
			assert.Equal(t, zapcore.InfoLevel, log.Level)
			assert.NotEmpty(t, log.Time)
			assert.Equal(t, tt.name, log.LoggerName)
			assert.False(t, log.Caller.Defined)

			fields := log.Context
			assert.Len(t, fields, 5)
			assert.Equal(t, tt.method, fieldByKey(fields, "method").String)
			assert.Equal(t, tt.path, fieldByKey(fields, "url").String)
			assert.Equal(t, int64(tt.status), fieldByKey(fields, "status").Integer)
			assert.Equal(t, int64(len(tt.body)), fieldByKey(fields, "size").Integer)
			assert.NotZero(t, fieldByKey(fields, "duration").Integer)
		})
	}
}

func fieldByKey(fields []zapcore.Field, key string) *zapcore.Field {
	for _, field := range fields {
		if field.Key == key {
			return &field
		}
	}
	return nil
}
