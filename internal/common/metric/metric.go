// Package metrics
// contains interface, implementations, factory and transformer for metrics
package metric

type Metric interface {
	Type() string        // Type of Metric
	Name() string        // Name of Metric
	StringValue() string // StringValue of Metric
	Clone() Metric       // Clone Metric
}
