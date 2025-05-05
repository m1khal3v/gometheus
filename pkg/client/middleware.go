package client

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type preRequestHook func(request *http.Request) error

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

func preRequestHookCombine(functions ...preRequestHook) resty.PreRequestHook {
	return func(client *resty.Client, request *http.Request) error {
		for _, function := range functions {
			if err := function(request); err != nil {
				return err
			}
		}

		return nil
	}
}

func (client *HTTPClient) encryptRequestBody(request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	_, err := io.Copy(buffer, request.Body)
	if err != nil {
		return err
	}

	cipher, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		client.config.publicKey,
		buffer.Bytes(),
	)
	if err != nil {
		return err
	}

	base64Encoded := make([]byte, base64.StdEncoding.EncodedLen(len(cipher)))
	base64.StdEncoding.Encode(base64Encoded, cipher)

	request.Body = io.NopCloser(bytes.NewBuffer(base64Encoded))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
	request.ContentLength = int64(buffer.Len())
	request.Header.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))
	request.Header.Set("Content-Encryption", "RSA-PKCS1v15")

	return nil
}

func (client *HTTPClient) compressRequestBody(request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	writer := client.gzipPool.Get().(*gzip.Writer)
	defer client.gzipPool.Put(writer)
	writer.Reset(buffer)

	_, err := io.Copy(writer, request.Body)
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

func (client *HTTPClient) addHMACSignature(request *http.Request) error {
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

	encoder := client.hmacPool.Get().(hash.Hash)
	defer client.hmacPool.Put(encoder)
	encoder.Reset()
	if _, err := encoder.Write(buffer.Bytes()); err != nil {
		return err
	}

	request.Header.Set(client.config.signature.header, hex.EncodeToString(encoder.Sum(nil)))

	return nil
}
