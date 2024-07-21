package middleware

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/m1khal3v/gometheus/internal/server/api"
	"hash"
	"io"
	"net/http"
	"sync"
)

type hmacPool struct {
	pool *sync.Pool
}

func newHMACPool(hash func() hash.Hash, key string) *hmacPool {
	return &hmacPool{
		pool: &sync.Pool{
			New: func() any {
				return hmac.New(hash, []byte(key))
			},
		},
	}
}

func HMACSignatureRespond(header string, hash func() hash.Hash, key string) func(next http.Handler) http.Handler {
	pool := newHMACPool(hash, key)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			wrapper := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			buffer := bytes.NewBuffer([]byte{})
			wrapper.Tee(buffer)
			wrapper.Discard() // disable writing to original writer

			next.ServeHTTP(wrapper, request)

			encoder, restore := pool.getHMAC()
			defer restore()

			writer.Header().Set("Content-Length", fmt.Sprintf("%d", wrapper.BytesWritten()))
			writer.Header().Set(header, hex.EncodeToString(encoder.Sum(nil)))
			writer.WriteHeader(wrapper.Status())

			if _, err := writer.Write(buffer.Bytes()); err != nil {
				api.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t write response", err)
			}
		})
	}
}

func HMACSignatureValidate(header string, hash func() hash.Hash, key string) func(next http.Handler) http.Handler {
	pool := newHMACPool(hash, key)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			signature := request.Header.Get(header)
			if signature == "" {
				api.WriteJSONErrorResponse(http.StatusBadRequest, writer, "Signature is not defined", nil)
				return
			}

			decodedSignature, err := hex.DecodeString(signature)
			if err != nil {
				api.WriteJSONErrorResponse(http.StatusBadRequest, writer, "Signature is not valid", err)
				return
			}

			buffer := bytes.NewBuffer([]byte{})
			if _, err := io.Copy(buffer, request.Body); err != nil {
				api.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t read request", err)
				return
			}

			request.Body = io.NopCloser(bytes.NewBuffer(buffer.Bytes()))

			encoder, restore := pool.getHMAC()
			defer restore()

			encoder.Write(buffer.Bytes())
			if !hmac.Equal(encoder.Sum(nil), decodedSignature) {
				api.WriteJSONErrorResponse(http.StatusBadRequest, writer, "Signature is not valid", nil)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}

func (hmacPool hmacPool) getHMAC() (hash.Hash, func()) {
	hmac := hmacPool.pool.Get().(hash.Hash)
	restore := func() {
		hmacPool.pool.Put(hmac)
	}
	hmac.Reset()

	return hmac, restore
}
