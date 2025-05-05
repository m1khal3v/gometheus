package middleware

import (
	"net"
	"net/http"

	"github.com/m1khal3v/gometheus/internal/server/api"
)

// SubnetValidate validate client subnet
func SubnetValidate(header string, subnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ipStr := request.Header.Get(header)
			if ipStr == "" {
				api.WriteJSONErrorResponse(http.StatusForbidden, writer, "IP header is missing", nil)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				api.WriteJSONErrorResponse(http.StatusForbidden, writer, "Invalid IP format", nil)
				return
			}

			if !subnet.Contains(ip) {
				api.WriteJSONErrorResponse(http.StatusForbidden, writer, "IP not in allowed subnet", nil)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}
