package api

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/logger"
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
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics := []metric.Metric{}
	for _, saveMetricRequest := range saveMetricsRequest {
		if _, err := govalidator.ValidateStruct(saveMetricRequest); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		metric, err := factory.NewFromRequest(saveMetricRequest)
		if err != nil {
			logger.Logger.Error(err.Error())
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		metrics = append(metrics, metric)
	}

	metrics, err := container.manager.SaveBatch(metrics)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	saveMetricsResponse := []responses.SaveMetricResponse{}
	for _, metric := range metrics {
		response, err := transformer.TransformToSaveResponse(metric)
		if err != nil {
			logger.Logger.Error(err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		saveMetricsResponse = append(saveMetricsResponse, *response)
	}

	jsonResponse, err := json.Marshal(saveMetricsResponse)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(jsonResponse)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
