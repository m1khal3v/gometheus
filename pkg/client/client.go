// Package client
// contains gometheus http client
package client

import (
	"context"

	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
)

type Client interface {
	SaveMetrics(ctx context.Context, requests []request.SaveMetricRequest) ([]response.SaveMetricResponse, *response.APIError, error)
}
