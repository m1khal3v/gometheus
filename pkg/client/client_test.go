package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

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
	client := New("test", WithoutRetry(), withTransport(function))

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
