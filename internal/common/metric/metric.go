package metric

type Metric interface {
	Type() string
	Name() string
	StringValue() string
	Clone() Metric
}
