// Package metrics
// contains interface, implementations, factory and transformer for metrics
package metric

type Metric interface {
	Type() string
	Name() string
	StringValue() string
	Clone() Metric
}
