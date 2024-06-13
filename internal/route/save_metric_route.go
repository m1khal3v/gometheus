package route

import (
	"github.com/m1khal3v/gometheus/internal/logger"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"net/http"
	"strings"
)

func (routeContainer Container) SaveMetric(writer http.ResponseWriter, request *http.Request) {
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

	metric, err := _metric.New(metricType, metricName, metricValue)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	current := routeContainer.Storage.Get(metricName)
	if current != nil {
		routeContainer.Storage.Save(_metric.Combine(current, metric))
	} else {
		routeContainer.Storage.Save(metric)
	}
	writer.WriteHeader(http.StatusOK)
}
