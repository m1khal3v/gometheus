package route

import (
	"fmt"
	"net/http"
)

func (routeContainer Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	metrics, err := routeContainer.Storage.GetAll()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	for _, metric := range metrics {
		_, _ = writer.Write([]byte(fmt.Sprintf("%s: %v\n", metric.GetName(), metric.GetStringValue())))
	}
}
