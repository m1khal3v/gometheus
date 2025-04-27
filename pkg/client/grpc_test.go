package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"sync"
	"testing"

	"github.com/m1khal3v/gometheus/pkg/proto"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type mockMetricsServiceClient struct {
	mock.Mock
	proto.MetricsServiceClient
}

func (m *mockMetricsServiceClient) SaveMetric(ctx context.Context, req *proto.SaveMetricRequest, opts ...grpc.CallOption) (*proto.SaveMetricResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.SaveMetricResponse), args.Error(1)
}

func (m *mockMetricsServiceClient) SaveMetrics(ctx context.Context, req *proto.SaveMetricsBatchRequest, opts ...grpc.CallOption) (*proto.SaveMetricsBatchResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.SaveMetricsBatchResponse), args.Error(1)
}

func TestGRPCClient_SaveMetric(t *testing.T) {
	mockClient := &mockMetricsServiceClient{}
	client := &GRPCClient{
		client: mockClient,
		config: newConfig("localhost"),
		hmacPool: &sync.Pool{
			New: func() any { return mockHasher() },
		},
	}

	req := &request.SaveMetricRequest{
		MetricName: "test_metric",
		MetricType: "counter",
		Delta:      ptrInt64(10),
		Value:      ptrFloat64(5.5),
	}
	grpcReq := &proto.SaveMetricRequest{
		MetricName: req.MetricName,
		MetricType: req.MetricType,
		Delta:      wrapperspb.Int64(10),
		Value:      wrapperspb.Double(5.5),
	}
	grpcResp := &proto.SaveMetricResponse{
		MetricName: "test_metric",
		MetricType: "counter",
		Delta:      wrapperspb.Int64(10),
		Value:      wrapperspb.Double(5.5),
	}

	mockClient.On("SaveMetric", mock.Anything, grpcReq).Return(grpcResp, nil)

	ctx := context.Background()
	resp, apiErr, err := client.SaveMetric(ctx, req)
	assert.NoError(t, err)
	assert.Nil(t, apiErr)
	assert.NotNil(t, resp)
	assert.Equal(t, "test_metric", resp.MetricName)
	assert.Equal(t, "counter", resp.MetricType)
	assert.Equal(t, int64(10), *resp.Delta)
	assert.Equal(t, 5.5, *resp.Value)

	mockClient.AssertExpectations(t)
}

func TestGRPCClient_SaveMetrics(t *testing.T) {
	mockClient := &mockMetricsServiceClient{}
	client := &GRPCClient{
		client: mockClient,
		config: newConfig("localhost"),
		hmacPool: &sync.Pool{
			New: func() any { return mockHasher() },
		},
	}

	requests := []request.SaveMetricRequest{
		{
			MetricName: "metric_1",
			MetricType: "gauge",
			Value:      ptrFloat64(3.8),
		},
		{
			MetricName: "metric_2",
			MetricType: "counter",
			Delta:      ptrInt64(42),
		},
	}
	batchReq := &proto.SaveMetricsBatchRequest{
		Metrics: []*proto.SaveMetricRequest{
			{
				MetricName: "metric_1",
				MetricType: "gauge",
				Value:      wrapperspb.Double(3.8),
			},
			{
				MetricName: "metric_2",
				MetricType: "counter",
				Delta:      wrapperspb.Int64(42),
			},
		},
	}
	batchResp := &proto.SaveMetricsBatchResponse{
		Metrics: []*proto.SaveMetricResponse{
			{
				MetricName: "metric_1",
				MetricType: "gauge",
				Value:      wrapperspb.Double(3.8),
			},
			{
				MetricName: "metric_2",
				MetricType: "counter",
				Delta:      wrapperspb.Int64(42),
			},
		},
	}

	mockClient.On("SaveMetrics", mock.Anything, batchReq).Return(batchResp, nil)

	ctx := context.Background()
	resp, apiErr, err := client.SaveMetrics(ctx, requests)
	assert.NoError(t, err)
	assert.Nil(t, apiErr)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "metric_1", resp[0].MetricName)
	assert.Equal(t, "gauge", resp[0].MetricType)
	assert.Equal(t, 3.8, *resp[0].Value)

	mockClient.AssertExpectations(t)
}

func ptrInt64(v int64) *int64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func mockHasher() hash.Hash {
	return hmac.New(func() hash.Hash { return sha256.New() }, []byte("dummy_key"))
}
