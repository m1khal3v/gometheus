package metric

import (
	"github.com/m1khal3v/gometheus/internal/metric/counter"
	"github.com/m1khal3v/gometheus/internal/metric/gauge"
)

func Combine(metric1 Metric, metric2 Metric) Metric {
	if metric1.GetType() != metric2.GetType() || metric1.GetName() != metric2.GetName() {
		return metric2
	}

	switch metric1.GetType() {
	case gauge.Type:
		return metric2
	case counter.Type:
		return metric2.(*counter.Metric).Add(metric1.(*counter.Metric))
	}

	return nil
}
