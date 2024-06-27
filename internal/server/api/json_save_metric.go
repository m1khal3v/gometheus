package api

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	"net/http"
)

func (container Container) JSONSaveMetric(writer http.ResponseWriter, request *http.Request) {
	saveMetricRequest := requests.SaveMetricRequest{}
	if err := json.NewDecoder(request.Body).Decode(&saveMetricRequest); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

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

	metric = container.manager.Save(metric)
	response, err := transformer.TransformToSaveResponse(metric)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(response)
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
