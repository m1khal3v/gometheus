package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gometheus/internal/store"
	"net/http"
)

type Client struct {
	resty *resty.Client
}

type UnexpectedStatusError struct {
	Status int
}

func (err UnexpectedStatusError) Error() string {
	return fmt.Sprintf("Unexpected status code: %v", err.Status)
}

func newUnexpectedStatusError(status int) UnexpectedStatusError {
	return UnexpectedStatusError{
		Status: status,
	}
}

func NewClient(endpoint string) *Client {
	return &Client{
		resty: resty.New().SetBaseURL(fmt.Sprintf("http://%s/", endpoint)),
	}
}

func (client *Client) SendMetric(metric *store.Metric) error {
	response, err := client.resty.R().SetPathParams(map[string]string{
		"type":  metric.Type,
		"name":  metric.Name,
		"value": metric.GetStringValue(),
	}).Post("update/{type}/{name}/{value}")

	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return newUnexpectedStatusError(response.StatusCode())
	}

	return nil
}
