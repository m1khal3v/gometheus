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
					reader, err := gzip.NewReader(bytes.NewReader([]byte{31, 139, 8, 0, 0, 0, 0, 0, 0,
						255, 203, 72, 205, 201, 201, 7, 0, 134, 166, 16, 54, 5, 0, 0, 0}))
					if err != nil {
						return nil
					}

					return reader
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
	Reset(r io.ReadCloser) error
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
	if gzipReader, ok := decoder.(gzipResetter); ok {
		err := gzipReader.Reset(request.Body)
		if err != nil {
			return nil, nil
		}
	}
	if deflateReader, ok := decoder.(flate.Resetter); ok {
		err := deflateReader.Reset(request.Body, nil)
		if err != nil {
			return nil, nil
		}
	}

	return decoder.(io.ReadCloser), restore
}
