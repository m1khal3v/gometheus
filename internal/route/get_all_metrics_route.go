package route

import (
	"fmt"
	"net/http"
)

func (routeContainer Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	for _, metric := range routeContainer.Storage.GetAll() {
		_, _ = writer.Write([]byte(fmt.Sprintf("%s: %s\n", metric.GetName(), metric.GetStringValue())))
	}
}
