// Package factory
// contains metrics factory
package factory

import (
	"fmt"
	"strconv"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/proto"
	"github.com/m1khal3v/gometheus/pkg/request"
)

type UnknownTypeError struct {
	Type string
}

func (err UnknownTypeError) Error() string {
	return fmt.Sprintf("metric type '%s' is not defined", err.Type)
}

func newErrUnknownType(metricType string) error {
	return &UnknownTypeError{
		Type: metricType,
	}
}

type InvalidValueError struct {
	Value string
}

func (err InvalidValueError) Error() string {
	return fmt.Sprintf("metric value '%s' is invalid", err.Value)
}

func newErrInvalidValue(value string) error {
	return &InvalidValueError{
		Value: value,
	}
}

func New(metricType, name, value string) (metric.Metric, error) {
	switch metricType {
	case gauge.MetricType:
		metricConvertedValue, err := strconv.ParseFloat(value, 64)
		if nil != err {
			return nil, newErrInvalidValue(value)
		}

		return gauge.New(name, metricConvertedValue), nil
	case counter.MetricType:
		metricConvertedValue, err := strconv.ParseInt(value, 10, 64)
		if nil != err {
			return nil, newErrInvalidValue(value)
		}

		return counter.New(name, metricConvertedValue), nil
	default:
		return nil, newErrUnknownType(metricType)
	}
}

func NewFromRequest(request *request.SaveMetricRequest) (metric.Metric, error) {
	switch request.MetricType {
	case gauge.MetricType:
		if nil == request.Value {
			return nil, newErrInvalidValue("nil")
		}

		return gauge.New(request.MetricName, *request.Value), nil
	case counter.MetricType:
		if nil == request.Delta {
			return nil, newErrInvalidValue("nil")
		}

		return counter.New(request.MetricName, *request.Delta), nil
	default:
		return nil, newErrUnknownType(request.MetricType)
	}
}

func NewFromGRPCRequest(request *proto.SaveMetricRequest) (metric.Metric, error) {
	switch request.MetricType {
	case gauge.MetricType:
		if nil == request.Value {
			return nil, newErrInvalidValue("nil")
		}

		return gauge.New(request.MetricName, request.Value.Value), nil
	case counter.MetricType:
		if nil == request.Delta {
			return nil, newErrInvalidValue("nil")
		}

		return counter.New(request.MetricName, request.Delta.Value), nil
	default:
		return nil, newErrUnknownType(request.MetricType)
	}
}
