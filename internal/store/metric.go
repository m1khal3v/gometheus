package store

import (
	"fmt"
	"slices"
	"strconv"
)

const (
	MetricTypeGauge   string = "gauge"
	MetricTypeCounter string = "counter"
)

type Metric struct {
	Type       string
	Name       string
	FloatValue float64
	IntValue   int64
}

type UnknownTypeError struct {
	Type string
}

func (err UnknownTypeError) Error() string {
	return fmt.Sprintf("Metric type '%v' is not defined", err.Type)
}

func newUnknownTypeError(metricType string) UnknownTypeError {
	return UnknownTypeError{
		Type: metricType,
	}
}

type InvalidValueError struct {
	Value string
}

func (err InvalidValueError) Error() string {
	return fmt.Sprintf("Metric value '%v' is invalid", err.Value)
}

func newInvalidValueError(value string) InvalidValueError {
	return InvalidValueError{
		Value: value,
	}
}

type InvalidValueTypeError struct{}

func (err InvalidValueTypeError) Error() string {
	return "Value type is not supported"
}

func resolveFloat64Value(metricType string, name string, value float64) (*Metric, error) {
	if metricType != MetricTypeGauge {
		return nil, newInvalidValueError(strconv.FormatFloat(value, 'f', -1, 64))
	}

	return &Metric{
		Type:       MetricTypeGauge,
		Name:       name,
		FloatValue: value,
	}, nil
}

func resolveInt64Value(metricType string, name string, value int64) (*Metric, error) {
	if metricType != MetricTypeCounter {
		return nil, newInvalidValueError(strconv.FormatInt(value, 10))
	}

	return &Metric{
		Type:     MetricTypeCounter,
		Name:     name,
		IntValue: value,
	}, nil
}

func resolveStringValue(metricType string, name string, value string) (*Metric, error) {
	if value == "" {
		return nil, newInvalidValueError(value)
	}

	switch metricType {
	case MetricTypeGauge:
		metricConvertedValue, err := strconv.ParseFloat(value, 64)
		if nil != err {
			return nil, newInvalidValueError(value)
		}

		return &Metric{
			Type:       MetricTypeGauge,
			Name:       name,
			FloatValue: metricConvertedValue,
		}, nil
	case MetricTypeCounter:
		metricConvertedValue, err := strconv.ParseInt(value, 10, 64)
		if nil != err {
			return nil, newInvalidValueError(value)
		}

		return &Metric{
			Type:     MetricTypeCounter,
			Name:     name,
			IntValue: metricConvertedValue,
		}, nil
	default:
		return nil, nil
	}
}

func NewMetric(metricType string, name string, value any) (*Metric, error) {
	err := ValidateMetricType(metricType)
	if nil != err {
		return nil, err
	}

	switch typeValue := value.(type) {
	case float64:
		return resolveFloat64Value(metricType, name, typeValue)
	case int64:
		return resolveInt64Value(metricType, name, typeValue)
	case string:
		return resolveStringValue(metricType, name, typeValue)
	default:
		return nil, InvalidValueTypeError{}
	}
}

func ValidateMetricType(metricType string) error {
	metricTypes := []string{
		MetricTypeGauge,
		MetricTypeCounter,
	}

	if !slices.Contains(metricTypes, metricType) {
		return newUnknownTypeError(metricType)
	}

	return nil
}

func (metric *Metric) GetValue() any {
	switch metric.Type {
	case MetricTypeGauge:
		return metric.FloatValue
	case MetricTypeCounter:
		return metric.IntValue
	}

	return nil
}

func (metric *Metric) GetStringValue() string {
	return fmt.Sprintf("%v", metric.GetValue())
}
