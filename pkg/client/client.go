package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
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

func (client *Client) SendMetric(metricType, metricName, metricValue string) error {
	response, err := client.resty.R().SetPathParams(map[string]string{
		"type":  metricType,
		"name":  metricName,
		"value": metricValue,
	}).Post("update/{type}/{name}/{value}")

	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return newUnexpectedStatusError(response.StatusCode())
	}

	return nil
}
