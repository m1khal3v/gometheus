package metric

type Metric interface {
	GetType() string
	GetName() string
	GetStringValue() string
	Replace(newMetric Metric) Metric
}
