package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
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

func New(endpoint string) *Client {
	return &Client{
		resty: resty.New().SetBaseURL(fmt.Sprintf("http://%s/", endpoint)),
	}
}

func (client *Client) SaveMetric(metricType, metricName, metricValue string) error {
	result, err := client.resty.R().SetPathParams(map[string]string{
		"type":  metricType,
		"name":  metricName,
		"value": metricValue,
	}).Post("update/{type}/{name}/{value}")
	if err != nil {
		return err
	}

	if result.StatusCode() != http.StatusOK {
		return newUnexpectedStatusError(result.StatusCode())
	}

	return nil
}

func (client *Client) SaveMetricAsJSON(request *request.SaveMetricRequest) (*response.SaveMetricResponse, error) {
	result, err := client.resty.R().SetBody(request).SetResult(&response.SaveMetricResponse{}).Post("update")
	if err != nil {
		return nil, err
	}

	if result.StatusCode() != http.StatusOK {
		return nil, newUnexpectedStatusError(result.StatusCode())
	}

	return result.Result().(*response.SaveMetricResponse), nil
}
