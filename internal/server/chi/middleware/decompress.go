package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

type decoderPool struct {
	pool map[string]*sync.Pool
}

func newDecoderPool() *decoderPool {
	return &decoderPool{
		pool: map[string]*sync.Pool{
			"gzip": {
				New: func() any {
					return new(gzip.Reader)
				},
			},
			"deflate": {
				New: func() any {
					return flate.NewReader(bytes.NewReader(nil))
				},
			},
		},
	}
}

func Decompress() func(next http.Handler) http.Handler {
	decoderPool := newDecoderPool()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			decoder, restore := decoderPool.getDecoder(request)
			if decoder != nil {
				defer decoder.Close()
				defer restore()

				request.Body = decoder
			}

			next.ServeHTTP(writer, request)
		})
	}
}

type gzipResetter interface {
	Reset(r io.Reader) error
}

func (decoderPool decoderPool) getDecoder(request *http.Request) (io.ReadCloser, func()) {
	encoding := request.Header.Get("Content-Encoding")
	pool, ok := decoderPool.pool[encoding]
	if !ok {
		return nil, nil
	}

	decoder := pool.Get()
	restore := func() {
		pool.Put(decoder)
	}

	switch encoding {
	case "gzip":
		err := decoder.(gzipResetter).Reset(request.Body)
		if err != nil {
			return nil, nil
		}
	case "deflate":
		err := decoder.(flate.Resetter).Reset(request.Body, nil)
		if err != nil {
			return nil, nil
		}
	}

	return decoder.(io.ReadCloser), restore
}
