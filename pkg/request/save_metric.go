package request

type SaveMetricRequest struct {
	MetricName string   `json:"id" valid:"required,minstringlength(1)"`
	MetricType string   `json:"type" valid:"required,minstringlength(1)"`
	Delta      *int64   `json:"delta"`
	Value      *float64 `json:"value"`
}
