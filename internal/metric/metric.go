package metric

type Metric interface {
	GetType() string
	GetName() string
	GetValue() string
}
