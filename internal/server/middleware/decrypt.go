package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/m1khal3v/gometheus/internal/server/api"
)

// Decrypt request body
func Decrypt(privKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			encryption := request.Header.Get("Content-Encryption")
			if encryption != "RSA-PKCS1v15" {
				next.ServeHTTP(writer, request)
				return
			}

			buffer := bytes.NewBuffer([]byte{})
			if _, err := io.Copy(buffer, request.Body); err != nil {
				api.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t read request", err)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(buffer.String())
			if err != nil {
				api.WriteJSONErrorResponse(http.StatusBadRequest, writer, "Can`t base64 decode request", err)
				return
			}

			decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, decoded)
			if err != nil {
				api.WriteJSONErrorResponse(http.StatusBadRequest, writer, "Can`t decrypt request", err)
				return
			}

			request.Body = io.NopCloser(bytes.NewReader(decrypted))
			request.ContentLength = int64(len(decrypted))
			request.Header.Del("Content-Encryption")

			next.ServeHTTP(writer, request)
		})
	}
}
