package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/m1khal3v/gometheus/pkg/retry"
	"io"
	"net/http"
	"time"
)

type Client struct {
	resty *resty.Client
}

type ErrUnexpectedStatus struct {
	Status int
}

func (err ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("Unexpected status code: %d", err.Status)
}

func newUnexpectedStatusError(status int) ErrUnexpectedStatus {
	return ErrUnexpectedStatus{
		Status: status,
	}
}

func New(endpoint string, compress bool) *Client {
	client := resty.
		New().
		SetBaseURL(fmt.Sprintf("http://%s/", endpoint)).
		SetHeader("Accept-Encoding", "gzip")

	if compress {
		client.SetPreRequestHook(compressRequestBody)
	}

	return &Client{resty: client}
}

func compressRequestBody(client *resty.Client, request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := new(bytes.Buffer)
	writer, err := gzip.NewWriterLevel(buffer, 5)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, request.Body)
	if err = errors.Join(err, writer.Close(), request.Body.Close()); err != nil {
		return err
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
	request.ContentLength = int64(buffer.Len())
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return nil
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
		if result == nil {
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
		if result == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*response.APIError), err
		}
	}

	return result.Result().(*response.SaveMetricResponse), nil, nil
}

func (client *Client) SaveMetricsAsJSON(ctx context.Context, requests []*request.SaveMetricRequest) ([]*response.SaveMetricResponse, *response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(requests).
		SetResult(&[]*response.SaveMetricResponse{}).
		SetError(&response.APIError{}),
		resty.MethodPost, "updates")

	if err != nil {
		if result == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*response.APIError), err
		}
	}

	return *result.Result().(*[]*response.SaveMetricResponse), nil, nil
}

func (client *Client) createRequest(ctx context.Context) *resty.Request {
	return client.resty.R().SetContext(ctx)
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	var result *resty.Response = nil
	err := retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		var err error
		result, err = request.Execute(method, url)
		if err != nil {
			return err
		}

		if result.StatusCode() != http.StatusOK {
			return newUnexpectedStatusError(result.StatusCode())
		}

		return nil
	}, func(err error) bool {
		return !errors.As(err, &ErrUnexpectedStatus{})
	})

	return result, err
}
