package factory

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/request"
	"strconv"
)

type ErrUnknownType struct {
	Type string
}

func (err ErrUnknownType) Error() string {
	return fmt.Sprintf("Metric type '%s' is not defined", err.Type)
}

func newUnknownTypeError(metricType string) ErrUnknownType {
	return ErrUnknownType{
		Type: metricType,
	}
}

type ErrInvalidValue struct {
	Value string
}

func (err ErrInvalidValue) Error() string {
	return fmt.Sprintf("Metric value '%s' is invalid", err.Value)
}

func newInvalidValueError(value string) ErrInvalidValue {
	return ErrInvalidValue{
		Value: value,
	}
}

func New(metricType, name, value string) (metric.Metric, error) {
	switch metricType {
	case gauge.Type:
		metricConvertedValue, err := strconv.ParseFloat(value, 64)
		if nil != err {
			return nil, newInvalidValueError(value)
		}

		return gauge.New(name, metricConvertedValue), nil
	case counter.Type:
		metricConvertedValue, err := strconv.ParseInt(value, 10, 64)
		if nil != err {
			return nil, newInvalidValueError(value)
		}

		return counter.New(name, metricConvertedValue), nil
	default:
		return nil, newUnknownTypeError(metricType)
	}
}

func NewFromRequest(request request.SaveMetricRequest) (metric.Metric, error) {
	switch request.MetricType {
	case gauge.Type:
		if nil == request.Value {
			return nil, newInvalidValueError("nil")
		}

		return gauge.New(request.MetricName, *request.Value), nil
	case counter.Type:
		if nil == request.Delta {
			return nil, newInvalidValueError("nil")
		}

		return counter.New(request.MetricName, *request.Delta), nil
	default:
		return nil, newUnknownTypeError(request.MetricType)
	}
}
