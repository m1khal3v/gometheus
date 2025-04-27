package api

import (
	"context"
	"net/http"

	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/pkg/proto"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func (container Container) JSONSaveMetric(writer http.ResponseWriter, request *http.Request) {
	saveMetricRequest, ok := DecodeAndValidateJSONRequest[requests.SaveMetricRequest](request, writer)
	if !ok {
		return
	}

	metric, err := factory.NewFromRequest(saveMetricRequest)
	if err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid metric data received", err)
		return
	}

	metric, err = container.manager.Save(request.Context(), metric)
	if err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t save metric", err)
		return
	}

	response, err := transformer.TransformToSaveResponse(metric)
	if err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t create response", err)
		return
	}

	WriteJSONResponse(response, writer)
}

func (container Container) SaveMetricInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		saveMetricRequest, ok := req.(*proto.SaveMetricRequest)
		if !ok {
			return handler(ctx, req)
		}

		metric, err := factory.NewFromGRPCRequest(saveMetricRequest)
		if err != nil {
			st := status.Newf(status.Code(err), "Invalid metric data: %v", err.Error())

			return nil, st.Err()
		}

		metric, err = container.manager.Save(ctx, metric)
		if err != nil {
			st := status.Newf(status.Code(err), "Failed to save metric: %v", err.Error())

			return nil, st.Err()
		}

		response, err := transformer.TransformToGRPCSaveResponse(metric)
		if err != nil {
			st := status.Newf(status.Code(err), "Failed to create response: %v", err.Error())

			return nil, st.Err()
		}

		return response, nil
	}
}
