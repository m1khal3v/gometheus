package api

import (
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	responses "github.com/m1khal3v/gometheus/pkg/response"
	"net/http"
)

func (container Container) JSONSaveMetrics(writer http.ResponseWriter, request *http.Request) {
	saveMetricsRequest := []requests.SaveMetricRequest{}
	if err := json.NewDecoder(request.Body).Decode(&saveMetricsRequest); err != nil {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return
	}

	if len(saveMetricsRequest) == 0 {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Empty request received", nil)
		return
	}

	var metrics []metric.Metric
	var errs []error
	for _, saveMetricRequest := range saveMetricsRequest {
		if _, err := govalidator.ValidateStruct(saveMetricRequest); err != nil {
			errs = append(errs, err)
			return
		}

		metric, err := factory.NewFromRequest(saveMetricRequest)
		if err != nil {
			errs = append(errs, err)
			return
		}

		metrics = append(metrics, metric)
	}

	if len(errs) > 0 {
		container.writeErrorResponse(http.StatusBadRequest, writer, "Invalid request received", errors.Join(errs...))
		return
	}

	metrics, err := container.manager.SaveBatch(metrics)
	if err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t save metrics", err)
		return
	}

	saveMetricsResponse := []responses.SaveMetricResponse{}
	for _, metric := range metrics {
		response, err := transformer.TransformToSaveResponse(metric)
		if err != nil {
			container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t create response", err)
			return
		}

		saveMetricsResponse = append(saveMetricsResponse, *response)
	}

	jsonResponse, err := json.Marshal(saveMetricsResponse)
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
