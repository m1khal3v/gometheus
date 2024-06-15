package api

import (
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/metric/factory"
	"net/http"
	"strings"
)

func (container Container) SaveMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")
	metricValue := request.PathValue("value")

	if strings.TrimSpace(metricType) == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(metricName) == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	metric, err := factory.New(metricType, metricName, metricValue)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	current := container.storage.Get(metricName)
	if current != nil {
		container.storage.Save(current.Replace(metric))
	} else {
		container.storage.Save(metric)
	}
	writer.WriteHeader(http.StatusOK)
}
