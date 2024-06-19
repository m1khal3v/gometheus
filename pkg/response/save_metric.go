package response

type SaveMetricResponse struct {
	MetricName string  `json:"id"`
	MetricType string  `json:"type"`
	Delta      int64   `json:"delta,omitempty"`
	Value      float64 `json:"value,omitempty"`
}
