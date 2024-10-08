package middleware

import (
	"fmt"
	"net/http"

	"github.com/m1khal3v/gometheus/internal/server/api"
)

// Recover panic in handlers and transform it to JSON error
func Recover() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				recovered := recover()

				switch recovered {
				case nil:
					return
				case http.ErrAbortHandler:
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(recovered)
				}

				err, ok := recovered.(error)
				if !ok {
					err = fmt.Errorf("%v", recovered)
				}

				api.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Internal server error", err)
			}()

			next.ServeHTTP(writer, request)
		})
	}
}
