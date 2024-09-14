// Package client
// contains gometheus http client
package client

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/m1khal3v/gometheus/pkg/retry"
)

type Client struct {
	resty  *resty.Client
	config *config
}

type ErrUnexpectedStatus struct {
	Status int
}

func (err ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.Status)
}

func newErrUnexpectedStatus(status int) ErrUnexpectedStatus {
	return ErrUnexpectedStatus{
		Status: status,
	}
}

var ErrInvalidSignature = errors.New("invalid Signature")

func New(address string, options ...ConfigOption) *Client {
	config := newConfig(address, options...)
	client := resty.
		New().
		SetTransport(config.transport).
		SetBaseURL(config.baseURL.String()).
		SetHeader("Accept-Encoding", "gzip")

	hooks := make([]preRequestHook, 0)
	if config.compress {
		hooks = append(hooks, compressRequestBody)
	}
	if config.signature != nil && config.signature.signRequest {
		hooks = append(hooks, addHMACSignature)
	}

	client.SetPreRequestHook(preRequestHookCombine(config, hooks...))

	return &Client{resty: client, config: config}
}

func (client *Client) SaveMetric(ctx context.Context, metricType, metricName, metricValue string) (*response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"type":  metricType,
			"name":  metricName,
			"value": metricValue,
		}).
		SetError(&response.APIError{}),
		resty.MethodPost, "update/{type}/{name}/{value}")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, err
		} else {
			return result.Error().(*response.APIError), err
		}
	}

	return nil, nil
}

func (client *Client) SaveMetricAsJSON(ctx context.Context, request *request.SaveMetricRequest) (*response.SaveMetricResponse, *response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetResult(&response.SaveMetricResponse{}).
		SetError(&response.APIError{}),
		resty.MethodPost, "update")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*response.APIError), err
		}
	}

	return result.Result().(*response.SaveMetricResponse), nil, nil
}

func (client *Client) SaveMetricsAsJSON(ctx context.Context, requests []request.SaveMetricRequest) ([]response.SaveMetricResponse, *response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(requests).
		SetResult(&[]response.SaveMetricResponse{}).
		SetError(&response.APIError{}),
		resty.MethodPost, "updates")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*response.APIError), err
		}
	}

	return *result.Result().(*[]response.SaveMetricResponse), nil, nil
}

func (client *Client) createRequest(ctx context.Context) *resty.Request {
	return client.resty.R().SetContext(ctx)
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	var result *resty.Response = nil
	do := func() error {
		var err error
		result, err = request.Execute(method, url)
		if err != nil {
			return err
		}

		if result.StatusCode() != http.StatusOK {
			return newErrUnexpectedStatus(result.StatusCode())
		}

		return nil
	}

	var err error
	if client.config.retry {
		err = retry.Retry(time.Second, 5*time.Second, 4, 2, do, func(err error) bool {
			return !errors.As(err, &ErrUnexpectedStatus{}) &&
				!errors.Is(err, context.DeadlineExceeded) &&
				!errors.Is(err, context.Canceled)
		})
	} else {
		err = do()
	}

	if err != nil {
		return result, err
	}

	signConfig := client.config.signature
	if signConfig != nil && signConfig.validateResponse {
		body, err := result.RawResponse.Body.(*BufferReader).ReadAll()
		if err != nil {
			return nil, err
		}

		resultSignature, err := hex.DecodeString(result.Header().Get(signConfig.header))
		if err != nil {
			return nil, err
		}

		if err := validateHMACSignature(body, resultSignature, []byte(signConfig.key), signConfig.hasher); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func validateHMACSignature(body, signature, key []byte, hash func() hash.Hash) error {
	encoder := hmac.New(hash, key)
	if _, err := encoder.Write(body); err != nil {
		return err
	}

	expectedSignature := encoder.Sum(nil)

	if !hmac.Equal(expectedSignature, signature) {
		return ErrInvalidSignature
	}

	return nil
}
