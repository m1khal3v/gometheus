package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"k8s.io/utils/ptr"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		wantErr       error
		wantBaseURL   string
		wantTransport http.RoundTripper
	}{
		{
			name: "valid only address",
			config: &Config{
				Address: "http://localhost",
			},
			wantBaseURL:   "http://localhost",
			wantTransport: http.DefaultTransport,
		},
		{
			name: "valid address + compress",
			config: &Config{
				Address:         "127.0.0.1:8080",
				DisableCompress: false,
			},
			wantBaseURL:   "http://127.0.0.1:8080",
			wantTransport: http.DefaultTransport,
		},
		{
			name: "valid address + compress + transport #1",
			config: &Config{
				Address:         "https://my.server.ru:443/api/",
				DisableCompress: true,
				transport:       &http.Transport{MaxIdleConns: 123},
			},
			wantBaseURL:   "https://my.server.ru:443/api",
			wantTransport: &http.Transport{MaxIdleConns: 123},
		},
		{
			name: "valid address + compress + transport #2",
			config: &Config{
				Address:         "https://my.server.ru:443/api/sub",
				DisableCompress: true,
				transport:       &http.Transport{MaxIdleConns: 123},
			},
			wantBaseURL:   "https://my.server.ru:443/api/sub",
			wantTransport: &http.Transport{MaxIdleConns: 123},
		},
		{
			name: "invalid address #1",
			config: &Config{
				Address: "ftp://localhost",
			},
			wantErr: newErrInvalidAddress("http://ftp://localhost"),
		},
		{
			name: "invalid address #2",
			config: &Config{
				Address: "http://ftp://localhost",
			},
			wantErr: newErrInvalidAddress("http://ftp://localhost"),
		},
		{
			name: "invalid address #3",
			config: &Config{
				Address: "https://my.server.ru:443/api?a=b&c=d",
			},
			wantErr: newErrInvalidAddress("https://my.server.ru:443/api?a=b&c=d"),
		},
		{
			name: "invalid address #4",
			config: &Config{
				Address: "https://127.0.0.1/api#fragment",
			},
			wantErr: newErrInvalidAddress("https://127.0.0.1/api#fragment"),
		},
		{
			name: "invalid address #5 disable check",
			config: &Config{
				Address:                  "https://localhost/api?param=1&param2=2",
				DisableAddressValidation: true,
			},
			wantBaseURL:   "https://localhost/api?param=1&param2=2",
			wantTransport: http.DefaultTransport,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.config)
			if tt.wantErr != nil {
				assert.Equal(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantBaseURL, got.resty.BaseURL)
				transport, err := got.resty.Transport()
				require.NoError(t, err)
				assert.Equal(t, tt.wantTransport, transport)
			}
		})
	}
}

func TestClient_SaveMetric(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		transport   roundTripFunction
		wantAPIErr  *response.APIError
		wantErr     error
	}{
		{
			name:        "valid",
			metricType:  "counter",
			metricName:  "test_metric",
			metricValue: "1",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
				}, nil
			}),
		},
		{
			name:        "api error",
			metricType:  "counter",
			metricName:  "test_metric",
			metricValue: "1",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, response.APIError{
					Code:    http.StatusBadRequest,
					Message: "Bad Request",
					Details: []string{
						"Invalid request",
					},
				}), nil
			}),
			wantErr: newErrUnexpectedStatus(http.StatusBadRequest),
			wantAPIErr: &response.APIError{
				Code:    http.StatusBadRequest,
				Message: "Bad Request",
				Details: []string{
					"Invalid request",
				},
			},
		},
		{
			name:        "transport error",
			metricType:  "counter",
			metricName:  "test_metric",
			metricValue: "1",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("test error")
			}),
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))
				assert.Equal(t, fmt.Sprintf("/update/%s/%s/%s", tt.metricType, tt.metricName, tt.metricValue), req.URL.Path)

				return tt.transport(req)
			})
			apiErr, err := client.SaveMetric(ctx, tt.metricType, tt.metricName, tt.metricValue)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}
		})
	}
}

func TestClient_SaveMetricAsJSON(t *testing.T) {
	tests := []struct {
		name       string
		request    *request.SaveMetricRequest
		transport  roundTripFunction
		want       *response.SaveMetricResponse
		wantAPIErr *response.APIError
		wantErr    error
	}{
		{
			name: "valid",
			request: &request.SaveMetricRequest{
				MetricName: "test_metric",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, response.SaveMetricResponse{
					MetricType: "counter",
					MetricName: "test_metric",
					Delta:      ptr.To(int64(123)),
				}), nil
			}),
			want: &response.SaveMetricResponse{
				MetricType: "counter",
				MetricName: "test_metric",
				Delta:      ptr.To(int64(123)),
			},
		},
		{
			name: "api error",
			request: &request.SaveMetricRequest{
				MetricName: "test_metric",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, response.APIError{
					Code:    http.StatusBadRequest,
					Message: "Bad Request",
					Details: []string{
						"Invalid request",
					},
				}), nil
			}),
			wantErr: newErrUnexpectedStatus(http.StatusBadRequest),
			wantAPIErr: &response.APIError{
				Code:    http.StatusBadRequest,
				Message: "Bad Request",
				Details: []string{
					"Invalid request",
				},
			},
		},
		{
			name: "transport error",
			request: &request.SaveMetricRequest{
				MetricName: "test_metric",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("test error")
			}),
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "/update", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.SaveMetricAsJSON(ctx, tt.request)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_SaveMetricsAsJSON(t *testing.T) {
	tests := []struct {
		name       string
		request    []request.SaveMetricRequest
		transport  roundTripFunction
		want       []response.SaveMetricResponse
		wantAPIErr *response.APIError
		wantErr    error
	}{
		{
			name: "valid",
			request: []request.SaveMetricRequest{{
				MetricName: "test_metric",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			}},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, []response.SaveMetricResponse{{
					MetricType: "counter",
					MetricName: "test_metric",
					Delta:      ptr.To(int64(123)),
				}}), nil
			}),
			want: []response.SaveMetricResponse{{
				MetricType: "counter",
				MetricName: "test_metric",
				Delta:      ptr.To(int64(123)),
			}},
		},
		{
			name: "api error",
			request: []request.SaveMetricRequest{{
				MetricName: "test_metric",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			}},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, response.APIError{
					Code:    http.StatusBadRequest,
					Message: "Bad Request",
					Details: []string{
						"Invalid request",
					},
				}), nil
			}),
			wantErr: newErrUnexpectedStatus(http.StatusBadRequest),
			wantAPIErr: &response.APIError{
				Code:    http.StatusBadRequest,
				Message: "Bad Request",
				Details: []string{
					"Invalid request",
				},
			},
		},
		{
			name: "transport error",
			request: []request.SaveMetricRequest{{
				MetricName: "test_metric",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			}},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("test error")
			}),
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "/updates", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.SaveMetricsAsJSON(ctx, tt.request)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

type roundTripFunction func(req *http.Request) (*http.Response, error)

func (function roundTripFunction) RoundTrip(req *http.Request) (*http.Response, error) {
	return function(req)
}

func newTestClient(t *testing.T, function roundTripFunction) *Client {
	t.Helper()
	client, err := New(&Config{
		DisableRetry:             true,
		DisableAddressValidation: true,
		transport:                function,
	})
	require.NoError(t, err)
	require.NotNil(t, client)

	return client
}

func createResponse(t *testing.T, statusCode int, response any) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: statusCode,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: createResponseBody(t, response),
	}
}

func createResponseBody(t *testing.T, response any) io.ReadCloser {
	t.Helper()
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	return io.NopCloser(bytes.NewReader(jsonResponse))
}
