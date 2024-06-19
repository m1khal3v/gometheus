package request

type GetMetricRequest struct {
	MetricName string `json:"id" valid:"required,minstringlength(1)"`
	MetricType string `json:"type" valid:"required,minstringlength(1)"`
}
