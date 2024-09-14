package client

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type preRequestHook func(config *Config, request *http.Request) error

type BufferReader struct {
	*bytes.Reader
}

func (buffer *BufferReader) Close() error {
	return nil
}

func (buffer *BufferReader) ReadAll() ([]byte, error) {
	if _, err := buffer.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return io.ReadAll(buffer)
}

type transportHook func(*http.Response) (*http.Response, error)

func (function transportHook) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return function(response)
}

var hmacTransport transportHook = func(response *http.Response) (*http.Response, error) {
	if response.Body == nil {
		return response, nil
	}

	buffer := bytes.NewBuffer([]byte{})

	_, err := io.Copy(buffer, response.Body)
	if err = errors.Join(err, response.Body.Close()); err != nil {
		return nil, err
	}

	response.Body = &BufferReader{Reader: bytes.NewReader(buffer.Bytes())}

	return response, nil
}

func preRequestHookCombine(config *Config, functions ...preRequestHook) resty.PreRequestHook {
	return func(client *resty.Client, request *http.Request) error {
		for _, function := range functions {
			if err := function(config, request); err != nil {
				return err
			}
		}

		return nil
	}
}

func compressRequestBody(config *Config, request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	writer, err := gzip.NewWriterLevel(buffer, 5)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, request.Body)
	if err = errors.Join(err, writer.Close(), request.Body.Close()); err != nil {
		return err
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
	request.ContentLength = int64(buffer.Len())
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return nil
}

func addHMACSignature(config *Config, request *http.Request) error {
	buffer := bytes.NewBuffer([]byte{})

	if request.Body != nil {
		_, err := io.Copy(buffer, request.Body)
		if err = errors.Join(err, request.Body.Close()); err != nil {
			return err
		}
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}

	encoder := hmac.New(config.Signature.Hasher, []byte(config.Signature.Key))
	if _, err := encoder.Write(buffer.Bytes()); err != nil {
		return err
	}

	request.Header.Set(config.Signature.Header, hex.EncodeToString(encoder.Sum(nil)))

	return nil
}