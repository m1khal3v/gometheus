package api

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	"net/http"
)

func (container Container) JSONSaveMetric(writer http.ResponseWriter, request *http.Request) {
	saveMetricRequest := requests.SaveMetricRequest{}
	if err := json.NewDecoder(request.Body).Decode(&saveMetricRequest); err != nil {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return
	}

	if _, err := govalidator.ValidateStruct(saveMetricRequest); err != nil {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid request received", err)
		return
	}

	metric, err := factory.NewFromRequest(saveMetricRequest)
	if err != nil {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid metric data received", err)
		return
	}

	metric, err = container.manager.Save(metric)
	if err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t save metric", err)
		return
	}

	response, err := transformer.TransformToSaveResponse(metric)
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
