package client

import (
	"compress/gzip"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"io"
	"net/http"
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
		client.SetPreRequestHook(func(client *resty.Client, request *http.Request) error {
			pipeReader, pipeWriter := io.Pipe()
			writer, err := gzip.NewWriterLevel(pipeWriter, 5)
			if err != nil {
				return err
			}

			_, err = io.Copy(writer, request.Body)
			if err != nil {
				return err
			}

			request.Body = pipeReader
			request.Header.Set("Content-Encoding", "gzip")
			request.Header.Del("Content-Length")

			return nil
		})
	}

	return &Client{
		resty: resty.
			New().
			SetBaseURL(fmt.Sprintf("http://%s/", endpoint)).
			SetHeader("Accept-Encoding", "gzip"),
	}
}

func (client *Client) SaveMetric(metricType, metricName, metricValue string) error {
	_, err := client.doRequest(client.
		resty.
		R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"type":  metricType,
			"name":  metricName,
			"value": metricValue,
		}),
		resty.MethodPost, "update/{type}/{name}/{value}")

	if err != nil {
		return err
	}

	return nil
}

func (client *Client) SaveMetricAsJSON(request *request.SaveMetricRequest) (*response.SaveMetricResponse, error) {
	result, err := client.doRequest(client.
		resty.
		R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetResult(&response.SaveMetricResponse{}),
		resty.MethodPost, "update")

	if err != nil {
		return nil, err
	}

	return result.Result().(*response.SaveMetricResponse), nil
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	result, err := request.Execute(method, url)
	if err != nil {
		return nil, err
	}

	if result.StatusCode() != http.StatusOK {
		return nil, newUnexpectedStatusError(result.StatusCode())
	}

	return result, nil
}
