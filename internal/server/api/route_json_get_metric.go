package api

import (
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	"net/http"
)

func (container Container) JSONGetMetric(writer http.ResponseWriter, request *http.Request) {
	getMetricRequest, ok := DecodeAndValidateJsonRequest[requests.GetMetricRequest](request, writer)
	if !ok {
		return
	}

	metric, err := container.manager.Get(request.Context(), getMetricRequest.MetricType, getMetricRequest.MetricName)
	switch {
	case err != nil:
		WriteJsonErrorResponse(http.StatusInternalServerError, writer, "Can`t get metric", err)
		return
	case metric == nil:
		WriteJsonErrorResponse(http.StatusNotFound, writer, "Metric not found", nil)
		return
	}

	response, err := transformer.TransformToGetResponse(metric)
	if err != nil {
		WriteJsonErrorResponse(http.StatusInternalServerError, writer, "Can`t create response", err)
		return
	}

	WriteJsonResponse(response, writer)
}
