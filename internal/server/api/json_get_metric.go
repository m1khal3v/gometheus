package api

import (
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric/transformer"
	"github.com/m1khal3v/gometheus/internal/server/manager"
	requests "github.com/m1khal3v/gometheus/pkg/request"
	"net/http"
)

func (container Container) JSONGetMetric(writer http.ResponseWriter, request *http.Request) {
	getMetricRequest := requests.GetMetricRequest{}
	if err := json.NewDecoder(request.Body).Decode(&getMetricRequest); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := govalidator.ValidateStruct(getMetricRequest); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := container.manager.Get(getMetricRequest.MetricType, getMetricRequest.MetricName)
	if err != nil {
		if errors.As(err, &manager.ErrMetricNotFound{}) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			logger.Logger.Error(err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	response, err := transformer.TransformToGetResponse(metric)
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
