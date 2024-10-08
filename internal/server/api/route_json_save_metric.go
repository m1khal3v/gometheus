package api

import (
	"net/http"

	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
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
