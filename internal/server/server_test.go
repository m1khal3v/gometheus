package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/internal/server/router"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	responses "github.com/m1khal3v/gometheus/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, server *httptest.Server, method string, path string, body []byte) (*http.Response, string) {
	request, err := http.NewRequest(method, server.URL+path, bytes.NewReader(body))
	require.NoError(t, err)

	response, err := server.Client().Do(request)
	require.NoError(t, err)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response, string(responseBody)
}

func TestSaveMetric(t *testing.T) {
	storage := memory.New()
	server := httptest.NewServer(router.New(storage, ""))
	defer server.Close()
	tests := []struct {
		method             string
		name               string
		metricType         string
		metricName         string
		metricValue        string
		previousValue      string
		expectedValue      string
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
			expectedBody:       "{\"code\":400,\"message\":\"Invalid metric data received\",\"details\":[\"metric value 'abc' is invalid\"]}",
		},
		{
			name:               "invalid counter",
			metricType:         "counter",
			metricName:         "test invalid counter",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "{\"code\":400,\"message\":\"Invalid metric data received\",\"details\":[\"metric value '123.321' is invalid\"]}",
		},
		{
			name:               "update gauge",
			metricType:         "gauge",
			metricName:         "test update gauge",
			metricValue:        "123.321",
			previousValue:      "456.654",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "update counter",
			metricType:         "counter",
			metricName:         "test update counter",
			metricValue:        "123",
			previousValue:      "321",
			expectedValue:      "444",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "empty type",
			metricType:         "",
			metricName:         "test empty type",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "{\"code\":400,\"message\":\"Empty type received\",\"details\":[]}",
		},
		{
			name:               "invalid type",
			metricType:         "test",
			metricName:         "test invalid type",
			metricValue:        "123.321",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "{\"code\":400,\"message\":\"Invalid metric data received\",\"details\":[\"metric type 'test' is not defined\"]}",
		},
		{
			name:               "empty name",
			metricType:         "counter",
			metricName:         "",
			metricValue:        "123",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "{\"code\":400,\"message\":\"Empty name received\",\"details\":[]}",
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
			ctx := context.Background()
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
			if tt.previousValue != "" {
				previousMetric, err := factory.New(tt.metricType, tt.metricName, tt.previousValue)
				require.NoError(t, err)
				require.NoError(t, storage.Save(ctx, previousMetric))
			}

			response, body := testRequest(t, server, method, path, nil)
			err := response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, body)
			if tt.expectedStatusCode == http.StatusOK {
				got, err := storage.Get(ctx, tt.metricName)
				require.NoError(t, err)
				if tt.expectedValue != "" {
					assert.Equal(t, tt.expectedValue, got.StringValue())
				} else {
					assert.Equal(t, tt.metricValue, got.StringValue())
				}
			}
		})
	}
}

type saveMetricRequest struct {
	MetricName any `json:"id"`
	MetricType any `json:"type"`
	Delta      any `json:"delta"`
	Value      any `json:"value"`
}

func TestSaveMetricJSON(t *testing.T) {
	storage := memory.New()
	server := httptest.NewServer(router.New(storage, ""))
	defer server.Close()
	tests := []struct {
		method             string
		name               string
		request            saveMetricRequest
		previous           metric.Metric
		expected           metric.Metric
		expectedStatusCode int
	}{
		{
			name: "valid gauge",
			request: saveMetricRequest{
				MetricName: "gauge",
				MetricType: "gauge",
				Value:      123.321,
			},
			expected:           gauge.New("gauge", 123.321),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "valid counter",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "counter",
				Delta:      123,
			},
			expected:           counter.New("counter", 123),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "update gauge",
			request: saveMetricRequest{
				MetricName: "gauge",
				MetricType: "gauge",
				Value:      123.321,
			},
			expected:           gauge.New("gauge", 123.321),
			previous:           gauge.New("gauge", 321.123),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "update counter",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "counter",
				Delta:      123,
			},
			previous:           counter.New("counter", 77),
			expected:           counter.New("counter", 200),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "invalid gauge field",
			request: saveMetricRequest{
				MetricName: "gauge",
				MetricType: "gauge",
				Delta:      123.321,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid counter field",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "counter",
				Value:      123,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid gauge value",
			request: saveMetricRequest{
				MetricName: "gauge",
				MetricType: "gauge",
				Value:      "abc",
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid counter value",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "counter",
				Delta:      123.321,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty gauge value",
			request: saveMetricRequest{
				MetricName: "gauge",
				MetricType: "gauge",
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty counter value",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "counter",
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty type",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "",
				Delta:      123,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty name",
			request: saveMetricRequest{
				MetricName: "",
				MetricType: "counter",
				Delta:      123,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid type",
			request: saveMetricRequest{
				MetricName: "counter",
				MetricType: "invalid",
				Delta:      123,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			method: http.MethodGet,
			name:   "invalid method",
			request: saveMetricRequest{
				MetricName: "gauge",
				MetricType: "gauge",
				Value:      123.321,
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			if tt.previous != nil {
				require.NoError(t, storage.Save(ctx, tt.previous))
			}

			bytes, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatal(err)
			}
			response, body := testRequest(t, server, method, "/update", bytes)
			err = response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			if tt.expectedStatusCode == http.StatusOK {
				expectedResponse, err := transformer.TransformToSaveResponse(tt.expected)
				require.NoError(t, err)
				expectedResponseBody, err := json.Marshal(expectedResponse)
				require.NoError(t, err)
				assert.Equal(t, string(expectedResponseBody), body)
			}
		})
	}
}

func TestSaveMetricsJSON(t *testing.T) {
	storage := memory.New()
	server := httptest.NewServer(router.New(storage, ""))
	defer server.Close()
	tests := []struct {
		method             string
		name               string
		request            []saveMetricRequest
		previous           []metric.Metric
		expected           []metric.Metric
		expectedStatusCode int
	}{
		{
			name: "valid gauge",
			request: []saveMetricRequest{
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Value:      123.321,
				},
				{
					MetricName: "gauge2",
					MetricType: "gauge",
					Value:      321.123,
				},
			},
			expected: []metric.Metric{
				gauge.New("gauge", 123.321),
				gauge.New("gauge2", 321.123),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "valid counter",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
					Delta:      123,
				},
				{
					MetricName: "counter2",
					MetricType: "counter",
					Delta:      321,
				},
			},
			expected: []metric.Metric{
				counter.New("counter", 123),
				counter.New("counter2", 321),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "valid multiple",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
					Delta:      123,
				},
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Value:      321.123,
				},
			},
			expected: []metric.Metric{
				counter.New("counter", 123),
				gauge.New("gauge", 321.123),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "update multiple",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
					Delta:      123,
				},
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Value:      321.123,
				},
			},
			expected: []metric.Metric{
				counter.New("counter", 444),
				gauge.New("gauge", 321.123),
			},
			previous: []metric.Metric{
				counter.New("counter", 321),
				gauge.New("gauge", 123.321),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "update multiple with collision",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
					Delta:      123,
				},
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Value:      321.123,
				},
				{
					MetricName: "counter",
					MetricType: "counter",
					Delta:      444,
				},
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Value:      333.333,
				},
			},
			expected: []metric.Metric{
				counter.New("counter", 888),
				gauge.New("gauge", 333.333),
			},
			previous: []metric.Metric{
				counter.New("counter", 321),
				gauge.New("gauge", 123.321),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "invalid gauge field",
			request: []saveMetricRequest{
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Delta:      123.321,
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid counter field",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
					Value:      123,
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid gauge value",
			request: []saveMetricRequest{
				{
					MetricName: "gauge",
					MetricType: "gauge",
					Value:      "abc",
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid counter value",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
					Delta:      123.321,
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty gauge value",
			request: []saveMetricRequest{
				{
					MetricName: "gauge",
					MetricType: "gauge",
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty counter value",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "counter",
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty type",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "",
					Delta:      123,
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty name",
			request: []saveMetricRequest{
				{
					MetricName: "",
					MetricType: "counter",
					Delta:      123,
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid type",
			request: []saveMetricRequest{
				{
					MetricName: "counter",
					MetricType: "invalid",
					Delta:      123,
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			method: http.MethodGet,
			name:   "invalid method",
			request: []saveMetricRequest{{
				MetricName: "gauge",
				MetricType: "gauge",
				Value:      123.321,
			}},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			if tt.previous != nil {
				for _, metric := range tt.previous {
					require.NoError(t, storage.Save(ctx, metric))
				}
			}

			bytes, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatal(err)
			}
			response, body := testRequest(t, server, method, "/updates", bytes)
			err = response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			if tt.expectedStatusCode == http.StatusOK {
				expectedResponses := make([]*responses.SaveMetricResponse, 0)
				for _, metric := range tt.expected {
					expectedResponse, err := transformer.TransformToSaveResponse(metric)
					require.NoError(t, err)
					expectedResponses = append(expectedResponses, expectedResponse)
				}
				actualResponses := make([]*responses.SaveMetricResponse, 0)
				require.NoError(t, json.Unmarshal([]byte(body), &actualResponses))
				assert.ElementsMatch(t, expectedResponses, actualResponses)
			}
			require.NoError(t, storage.Reset(ctx))
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		method             string
		name               string
		preset             map[string]metric.Metric
		metricType         string
		metricName         string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:       "valid gauge",
			metricType: "gauge",
			metricName: "test gauge",
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedBody:       "123.321",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "valid counter",
			metricType: "counter",
			metricName: "test counter",
			preset: map[string]metric.Metric{
				"test counter": counter.New("test counter", 123),
			},
			expectedBody:       "123",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "invalid gauge",
			metricType: "gauge",
			metricName: "test invalid gauge",
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"code\":404,\"message\":\"Metric not found\",\"details\":[]}",
		},
		{
			name:       "invalid counter",
			metricType: "counter",
			metricName: "test invalid counter",
			preset: map[string]metric.Metric{
				"test counter": counter.New("test counter", 123),
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"code\":404,\"message\":\"Metric not found\",\"details\":[]}",
		},
		{
			name:       "empty type",
			metricType: "",
			metricName: "test empty type",
			preset: map[string]metric.Metric{
				"test empty type": gauge.New("test empty type", 123.321),
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"code\":404,\"message\":\"Metric not found\",\"details\":[]}",
		},
		{
			name:       "invalid type",
			metricType: "test",
			metricName: "test invalid type",
			preset: map[string]metric.Metric{
				"test invalid type": gauge.New("test invalid type", 123.321),
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"code\":404,\"message\":\"Metric not found\",\"details\":[]}",
		},
		{
			name:       "empty name",
			metricType: "gauge",
			metricName: "",
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "404 page not found\n",
		},
		{
			method:     http.MethodPut,
			name:       "invalid method",
			metricType: "counter",
			metricName: "test invalid method",
			preset: map[string]metric.Metric{
				"test invalid method": counter.New("test invalid method", 123),
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := memory.New()
			server := httptest.NewServer(router.New(storage, ""))
			defer server.Close()

			for _, metric := range tt.preset {
				require.NoError(t, storage.Save(ctx, metric))
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

			response, body := testRequest(t, server, method, path, nil)
			err := response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

type getMetricRequest struct {
	MetricName string `json:"id"`
	MetricType string `json:"type"`
}

func TestGetMetricJSON(t *testing.T) {
	tests := []struct {
		method             string
		name               string
		preset             map[string]metric.Metric
		request            getMetricRequest
		expectedStatusCode int
		expected           metric.Metric
	}{
		{
			name: "valid gauge",
			request: getMetricRequest{
				MetricName: "test gauge",
				MetricType: "gauge",
			},
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expected:           gauge.New("test gauge", 123.321),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "valid counter",
			request: getMetricRequest{
				MetricName: "test counter",
				MetricType: "counter",
			},
			preset: map[string]metric.Metric{
				"test counter": counter.New("test counter", 123),
			},
			expected:           counter.New("test counter", 123),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "invalid gauge",
			request: getMetricRequest{
				MetricName: "test invalid gauge",
				MetricType: "gauge",
			},
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "invalid counter",
			request: getMetricRequest{
				MetricName: "test invalid counter",
				MetricType: "counter",
			},
			preset: map[string]metric.Metric{
				"test counter": counter.New("test counter", 123),
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "empty type",
			request: getMetricRequest{
				MetricName: "test empty type",
				MetricType: "",
			},
			preset: map[string]metric.Metric{
				"test empty type": gauge.New("test empty type", 123.321),
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid type",
			request: getMetricRequest{
				MetricName: "test invalid type",
				MetricType: "invalid",
			},
			preset: map[string]metric.Metric{
				"test invalid type": gauge.New("test invalid type", 123.321),
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "empty name",
			request: getMetricRequest{
				MetricName: "",
				MetricType: "gauge",
			},
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPut,
			name:   "invalid method",
			request: getMetricRequest{
				MetricName: "test gauge",
				MetricType: "gauge",
			},
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := memory.New()
			server := httptest.NewServer(router.New(storage, ""))
			defer server.Close()

			for _, metric := range tt.preset {
				require.NoError(t, storage.Save(ctx, metric))
			}

			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			bytes, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatal(err)
			}

			response, body := testRequest(t, server, method, "/value", bytes)
			err = response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			if tt.expectedStatusCode == http.StatusOK {
				expectedResponse, err := transformer.TransformToGetResponse(tt.expected)
				require.NoError(t, err)
				expectedResponseBody, err := json.Marshal(expectedResponse)
				require.NoError(t, err)
				assert.Equal(t, string(expectedResponseBody), body)
			}
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		method             string
		name               string
		preset             map[string]metric.Metric
		expectedStatusCode int
	}{
		{
			name:               "empty metrics",
			preset:             map[string]metric.Metric{},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "not empty metrics",
			preset: map[string]metric.Metric{
				"test gauge":          gauge.New("test gauge", 123.321),
				"test counter":        counter.New("test counter", 123),
				"test empty type":     gauge.New("test empty type", 123.321),
				"test invalid type":   gauge.New("test invalid type", 123.321),
				"test invalid method": counter.New("test invalid method", 123),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "single metric",
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			method: http.MethodPost,
			name:   "invalid method",
			preset: map[string]metric.Metric{
				"test gauge": gauge.New("test gauge", 123.321),
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := memory.New()
			server := httptest.NewServer(router.New(storage, ""))
			defer server.Close()

			for _, metric := range tt.preset {
				require.NoError(t, storage.Save(ctx, metric))
			}

			method := tt.method
			if method == "" {
				method = http.MethodGet
			}

			response, body := testRequest(t, server, method, "/", nil)
			err := response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			if tt.expectedStatusCode == http.StatusOK {
				for _, metric := range tt.preset {
					assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(
						"<tr>\\n +<td>%s<\\/td>\\n +<td>%s<\\/td>\\n +<td>%s<\\/td>\\n +<\\/tr>",
						regexp.QuoteMeta(metric.Name()),
						regexp.QuoteMeta(metric.Type()),
						regexp.QuoteMeta(metric.StringValue()),
					)), body)
				}
			}
		})
	}
}
