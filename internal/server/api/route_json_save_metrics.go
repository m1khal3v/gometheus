package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/pkg/proto"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	responses "github.com/m1khal3v/gometheus/pkg/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func (container Container) JSONSaveMetrics(writer http.ResponseWriter, request *http.Request) {
	saveMetricsRequest, ok := DecodeAndValidateJSONRequests[requests.SaveMetricRequest](request, writer)
	if !ok {
		return
	}

	var errs []error
	var metrics []metric.Metric
	for _, saveMetricRequest := range saveMetricsRequest {
		metric, err := factory.NewFromRequest(saveMetricRequest)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metrics = append(metrics, metric)
	}

	if len(errs) > 0 {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid request data received", errors.Join(errs...))
		return
	}

	metrics, err := container.manager.SaveBatch(request.Context(), metrics)
	if err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t save metrics", err)
		return
	}

	saveMetricsResponse := []responses.SaveMetricResponse{}
	for _, metric := range metrics {
		response, err := transformer.TransformToSaveResponse(metric)
		if err != nil {
			WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t create response", err)
			return
		}

		saveMetricsResponse = append(saveMetricsResponse, *response)
	}

	WriteJSONResponse(saveMetricsResponse, writer)
}

func (container Container) SaveMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		saveMetricsRequest, ok := req.(*proto.SaveMetricsBatchRequest)
		if !ok {
			return handler(ctx, req)
		}

		var metrics []metric.Metric
		var errs []error
		for _, saveMetricRequest := range saveMetricsRequest.Metrics {
			metric, err := factory.NewFromGRPCRequest(saveMetricRequest)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			metrics = append(metrics, metric)
		}

		if len(errs) > 0 {
			st := status.Newf(status.Code(errors.New("invalid request data")), "Invalid request data received: %v", errs)

			return nil, st.Err()
		}

		metrics, err := container.manager.SaveBatch(ctx, metrics)
		if err != nil {
			st := status.Newf(status.Code(err), "Failed to save metrics: %v", err.Error())

			return nil, st.Err()
		}

		var saveMetricsResponse proto.SaveMetricsBatchResponse
		for _, metric := range metrics {
			response, err := transformer.TransformToGRPCSaveResponse(metric)
			if err != nil {
				st := status.Newf(status.Code(err), "Failed to create response: %v", err.Error())
				return nil, st.Err()
			}

			saveMetricsResponse.Metrics = append(saveMetricsResponse.Metrics, response)
		}

		return &saveMetricsResponse, nil
	}
}
