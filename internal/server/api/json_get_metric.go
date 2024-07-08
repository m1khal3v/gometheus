package api

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	"net/http"
)

func (container Container) JSONGetMetric(writer http.ResponseWriter, request *http.Request) {
	getMetricRequest := requests.GetMetricRequest{}
	if err := json.NewDecoder(request.Body).Decode(&getMetricRequest); err != nil {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return
	}

	if _, err := govalidator.ValidateStruct(getMetricRequest); err != nil {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid request received", err)
		return
	}

	metric, err := container.manager.Get(getMetricRequest.MetricType, getMetricRequest.MetricName)
	if err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t get metric", err)
		return
	}
	if metric == nil {
		container.writeErrorResponse(http.StatusNotFound, writer, "Metric not found", nil)
		return
	}

	response, err := transformer.TransformToGetResponse(metric)
	if err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t create response", err)
		return
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t encode response", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if _, err = writer.Write(jsonResponse); err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t write response", err)
		return
	}
}
