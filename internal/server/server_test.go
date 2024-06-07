package server

import (
	"fmt"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/router"
	"github.com/m1khal3v/gometheus/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, server *httptest.Server, method string, path string) (*http.Response, string) {
	request, err := http.NewRequest(method, server.URL+path, nil)
	require.NoError(t, err)

	response, err := server.Client().Do(request)
	require.NoError(t, err)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response, string(responseBody)
}

func TestSaveMetric(t *testing.T) {
	server := httptest.NewServer(router.NewRouter(memory.NewStorage()))
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
			_ = response.Body.Close()
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		method             string
		name               string
		preset             map[string]*_metric.Metric
		metricType         string
		metricName         string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:       "valid gauge",
			metricType: "gauge",
			metricName: "test gauge",
			preset: map[string]*_metric.Metric{
				"test gauge": {
					Name:       "test gauge",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedBody:       "123.321",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "valid counter",
			metricType: "counter",
			metricName: "test counter",
			preset: map[string]*_metric.Metric{
				"test counter": {
					Name:     "test counter",
					Type:     "counter",
					IntValue: int64(123),
				},
			},
			expectedBody:       "123",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "invalid gauge",
			metricType: "gauge",
			metricName: "test invalid gauge",
			preset: map[string]*_metric.Metric{
				"test gauge": {
					Name:       "test gauge",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "invalid counter",
			metricType: "counter",
			metricName: "test invalid counter",
			preset: map[string]*_metric.Metric{
				"test counter": {
					Name:     "test counter",
					Type:     "counter",
					IntValue: int64(123),
				},
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "empty type",
			metricType: "",
			metricName: "test empty type",
			preset: map[string]*_metric.Metric{
				"test empty type": {
					Name:       "test empty type",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid type",
			metricType: "test",
			metricName: "test invalid type",
			preset: map[string]*_metric.Metric{
				"test invalid type": {
					Name:       "test invalid type",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "",
		},
		{
			name:       "empty name",
			metricType: "gauge",
			metricName: "",
			preset: map[string]*_metric.Metric{
				"test gauge": {
					Name:       "test gauge",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "404 page not found\n",
		},
		{
			method:     http.MethodPut,
			name:       "invalid method",
			metricType: "counter",
			metricName: "test invalid method",
			preset: map[string]*_metric.Metric{
				"test invalid method": {
					Name:     "test invalid method",
					Type:     "counter",
					IntValue: int64(123),
				},
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewStorage()
			server := httptest.NewServer(router.NewRouter(storage))
			defer server.Close()
			for _, metric := range tt.preset {
				_ = storage.Save(metric)
			}
			path := fmt.Sprintf(
				"/value/%v/%v",
				tt.metricType,
				tt.metricName,
			)
			method := tt.method
			if method == "" {
				method = http.MethodGet
			}
			response, body := testRequest(t, server, method, path)
			_ = response.Body.Close()
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		method             string
		name               string
		preset             map[string]*_metric.Metric
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "empty metrics",
			preset:             map[string]*_metric.Metric{},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "not empty metrics",
			preset: map[string]*_metric.Metric{
				"test gauge": {
					Name:       "test gauge",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
				"test counter": {
					Name:     "test counter",
					Type:     "counter",
					IntValue: int64(123),
				},
				"test empty type": {
					Name:       "test empty type",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
				"test invalid type": {
					Name:       "test invalid type",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
				"test invalid method": {
					Name:     "test invalid method",
					Type:     "counter",
					IntValue: int64(123),
				},
			},
			expectedBody:       "test invalid method: 123\ntest gauge: 123.321\ntest counter: 123\ntest empty type: 123.321\ntest invalid type: 123.321\n",
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "single metric",
			preset: map[string]*_metric.Metric{
				"test gauge": {
					Name:       "test gauge",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedBody:       "test gauge: 123.321\n",
			expectedStatusCode: http.StatusOK,
		},
		{
			method: http.MethodPost,
			name:   "invalid method",
			preset: map[string]*_metric.Metric{
				"test gauge": {
					Name:       "test gauge",
					Type:       "gauge",
					FloatValue: float64(123.321),
				},
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewStorage()
			server := httptest.NewServer(router.NewRouter(storage))
			defer server.Close()
			for _, metric := range tt.preset {
				_ = storage.Save(metric)
			}
			method := tt.method
			if method == "" {
				method = http.MethodGet
			}
			response, body := testRequest(t, server, method, "/")
			_ = response.Body.Close()
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			if !strings.Contains(tt.expectedBody, "\n") {
				assert.Equal(t, tt.expectedBody, body)
			} else {
				for _, part := range strings.Split(tt.expectedBody, "\n") {
					assert.Contains(t, body, part)
				}
			}
		})
	}
}
