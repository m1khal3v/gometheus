package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func ZapLogger(logger *zap.Logger, name string) func(next http.Handler) http.Handler {
	logger = logger.Named(name).WithOptions(zap.WithCaller(false))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			wrapper := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			timestamp := time.Now()
			defer func() {
				logger.Info(
					"Request processed",
					zap.String("method", request.Method),
					zap.String("url", request.URL.String()),
					zap.Int("status", wrapper.Status()),
					zap.Int("size", wrapper.BytesWritten()),
					zap.Duration("duration", time.Since(timestamp)),
				)
			}()
			next.ServeHTTP(wrapper, request)
		})
	}
}
