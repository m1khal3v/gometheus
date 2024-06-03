package client

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/storage"
	"net/http"
)

const host = "http://localhost:8080/"

func SendMetric(metric storage.Metric) error {
	response, err := http.Post(
		fmt.Sprintf(
			"%v/update/%v/%v/%v",
			host,
			metric.Type,
			metric.Name,
			metric.String(),
		),
		"text/plain",
		nil,
	)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		// TODO: когда будут нормальные ошибки можно будет сделать что-нибудь повеселее
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}
