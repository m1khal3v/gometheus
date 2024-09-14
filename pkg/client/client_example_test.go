package client

import (
	"context"
	"fmt"

	"github.com/m1khal3v/gometheus/pkg/request"
	"k8s.io/utils/ptr"
)

func Example() {
	client := New(
		"foo.bar.com",
		WithScheme("https"),
		WithPort(4433),
	)
	responses, apiErr, err := client.SaveMetricsAsJSON(context.TODO(), []request.SaveMetricRequest{
		{
			MetricName: "clicks",
			MetricType: "counter",
			Delta:      ptr.To[int64](10),
		},
		{
			MetricName: "views",
			MetricType: "counter",
			Delta:      ptr.To[int64](20),
		},
		{
			MetricName: "spent",
			MetricType: "gauge",
			Value:      ptr.To[float64](1024.512),
		},
	})
	if apiErr != nil {
		fmt.Printf("API error. Code: %d. Message: %s. Details: %v", apiErr.Code, apiErr.Message, apiErr.Details)
		return
	}
	if err != nil {
		fmt.Printf("Failed to save metrics: %s", err.Error())
		return
	}

	for _, response := range responses {
		var value any
		if response.MetricType == "gauge" {
			value = response.Value
		} else { // response.MetricType == "counter"
			value = response.Delta
		}

		fmt.Printf("Metric '%s' with type %s saved successfully. Current value: %v", response.MetricName, response.MetricType, value)
	}
}
