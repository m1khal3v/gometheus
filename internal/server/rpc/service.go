package rpc

import (
	"context"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/internal/server/manager"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsService struct {
	proto.UnimplementedMetricsServiceServer
	manager *manager.Manager
}

func NewMetricsService(storage storage.Storage) *MetricsService {
	return &MetricsService{manager: manager.New(storage)}
}

func (s *MetricsService) SaveMetric(
	ctx context.Context,
	req *proto.SaveMetricRequest,
) (*proto.SaveMetricResponse, error) {
	metric, err := factory.NewFromGRPCRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	savedMetric, err := s.manager.Save(ctx, metric)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp, err := transformer.TransformToGRPCSaveResponse(savedMetric)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *MetricsService) SaveMetrics(
	ctx context.Context,
	req *proto.SaveMetricsBatchRequest,
) (*proto.SaveMetricsBatchResponse, error) {
	var metrics []metric.Metric
	var errors []error

	for _, grpcReq := range req.GetMetrics() {
		m, err := factory.NewFromGRPCRequest(grpcReq)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		metrics = append(metrics, m)
	}

	if len(errors) > 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid metrics in batch: "+combineErrors(errors),
		)
	}

	savedMetrics, err := s.manager.SaveBatch(ctx, metrics)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var responses []*proto.SaveMetricResponse
	for _, m := range savedMetrics {
		resp, err := transformer.TransformToGRPCSaveResponse(m)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		responses = append(responses, resp)
	}

	return &proto.SaveMetricsBatchResponse{Metrics: responses}, nil
}

func combineErrors(errs []error) string {
	var result string
	for _, e := range errs {
		result += e.Error() + "; "
	}
	return result
}
