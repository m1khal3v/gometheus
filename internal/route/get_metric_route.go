package route

import (
	"net/http"
)

func (routeContainer Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	metric := routeContainer.Storage.Get(metricName)
	if metric == nil || metric.GetType() != metricType {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(metric.GetStringValue()))
}
