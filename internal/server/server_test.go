package server

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, server *httptest.Server, method string, path string) (*http.Response, string) {
	request, err := http.NewRequest(method, server.URL+path, nil)
	require.NoError(t, err)

	response, err := server.Client().Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response, string(responseBody)
}

func TestSaveMetric(t *testing.T) {
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()
	tests := []struct {
		method             string
		name               string
		metricType         string
		metricName         string
		metricValue        string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "valid gauge",
			metricType:         "gauge",
			metricName:         "test valid gauge",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "valid counter",
			metricType:         "counter",
			metricName:         "test valid counter",
			metricValue:        "123",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid gauge",
			metricType:         "gauge",
			metricName:         "test invalid gauge",
			metricValue:        "abc",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid counter",
			metricType:         "counter",
			metricName:         "test invalid counter",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty type",
			metricType:         "",
			metricName:         "test empty type",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid type",
			metricType:         "test",
			metricName:         "test invalid type",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty name",
			metricType:         "counter",
			metricName:         "",
			metricValue:        "123",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "empty value",
			metricType:         "counter",
			metricName:         "test empty value",
			metricValue:        "",
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "404 page not found\n",
		},
		{
			method:             http.MethodGet,
			name:               "invalid method",
			metricType:         "counter",
			metricName:         "test invalid method",
			metricValue:        "123",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf(
				"/update/%v/%v/%v",
				tt.metricType,
				tt.metricName,
				tt.metricValue,
			)
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			response, body := testRequest(t, server, method, path)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}
