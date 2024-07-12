package api

import (
	"errors"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	responses "github.com/m1khal3v/gometheus/pkg/response"
	"net/http"
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
