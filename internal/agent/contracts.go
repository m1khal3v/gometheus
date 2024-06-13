package agent

type metricSender interface {
	SendMetric(metricType, metricName, metricValue string) error
}
