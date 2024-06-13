package agent

type apiClient interface {
	SendMetric(metricType, metricName, metricValue string) error
}
