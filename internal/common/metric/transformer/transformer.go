package transformer

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
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

func TransformToSaveRequest(metric metric.Metric) (*request.SaveMetricRequest, error) {
	switch metric.GetType() {
	case gauge.Type:
		value := metric.(*gauge.Metric).GetValue()
		return &request.SaveMetricRequest{
			MetricType: metric.GetType(),
			MetricName: metric.GetName(),
			Value:      &value,
		}, nil
	case counter.Type:
		value := metric.(*counter.Metric).GetValue()
		return &request.SaveMetricRequest{
			MetricType: metric.GetType(),
			MetricName: metric.GetName(),
			Delta:      &value,
		}, nil
	}
	return nil, newUnknownTypeError(metric.GetType())
}

func TransformToSaveResponse(metric metric.Metric) (*response.SaveMetricResponse, error) {
	switch metric.GetType() {
	case gauge.Type:
		value := metric.(*gauge.Metric).GetValue()
		return &response.SaveMetricResponse{
			MetricType: metric.GetType(),
			MetricName: metric.GetName(),
			Value:      &value,
		}, nil
	case counter.Type:
		value := metric.(*counter.Metric).GetValue()
		return &response.SaveMetricResponse{
			MetricType: metric.GetType(),
			MetricName: metric.GetName(),
			Delta:      &value,
		}, nil
	}
	return nil, newUnknownTypeError(metric.GetType())
}

func TransformToGetResponse(metric metric.Metric) (*response.GetMetricResponse, error) {
	switch metric.GetType() {
	case gauge.Type:
		value := metric.(*gauge.Metric).GetValue()
		return &response.GetMetricResponse{
			MetricType: metric.GetType(),
			MetricName: metric.GetName(),
			Value:      &value,
		}, nil
	case counter.Type:
		value := metric.(*counter.Metric).GetValue()
		return &response.GetMetricResponse{
			MetricType: metric.GetType(),
			MetricName: metric.GetName(),
			Delta:      &value,
		}, nil
	}
	return nil, newUnknownTypeError(metric.GetType())
}
