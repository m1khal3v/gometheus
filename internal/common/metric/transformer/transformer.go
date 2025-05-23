// Package transformer
// contains metric -> request/response transformer
package transformer

import (
	"fmt"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/proto"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

func TransformToSaveRequest(metric metric.Metric) (*request.SaveMetricRequest, error) {
	switch metric.Type() {
	case gauge.MetricType:
		value := metric.(*gauge.Metric).GetValue()
		return &request.SaveMetricRequest{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Value:      &value,
		}, nil
	case counter.MetricType:
		value := metric.(*counter.Metric).GetValue()
		return &request.SaveMetricRequest{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Delta:      &value,
		}, nil
	}
	return nil, newErrUnknownType(metric.Type())
}

func TransformToSaveResponse(metric metric.Metric) (*response.SaveMetricResponse, error) {
	switch metric.Type() {
	case gauge.MetricType:
		value := metric.(*gauge.Metric).GetValue()
		return &response.SaveMetricResponse{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Value:      &value,
		}, nil
	case counter.MetricType:
		value := metric.(*counter.Metric).GetValue()
		return &response.SaveMetricResponse{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Delta:      &value,
		}, nil
	}
	return nil, newErrUnknownType(metric.Type())
}

func TransformToGetResponse(metric metric.Metric) (*response.GetMetricResponse, error) {
	switch metric.Type() {
	case gauge.MetricType:
		value := metric.(*gauge.Metric).GetValue()
		return &response.GetMetricResponse{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Value:      &value,
		}, nil
	case counter.MetricType:
		value := metric.(*counter.Metric).GetValue()
		return &response.GetMetricResponse{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Delta:      &value,
		}, nil
	}
	return nil, newErrUnknownType(metric.Type())
}

func TransformToGRPCSaveResponse(metric metric.Metric) (*proto.SaveMetricResponse, error) {
	switch metric.Type() {
	case gauge.MetricType:
		value := metric.(*gauge.Metric).GetValue()
		return &proto.SaveMetricResponse{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Value:      wrapperspb.Double(value),
		}, nil
	case counter.MetricType:
		value := metric.(*counter.Metric).GetValue()
		return &proto.SaveMetricResponse{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			Delta:      wrapperspb.Int64(value),
		}, nil
	}
	return nil, newErrUnknownType(metric.Type())
}
