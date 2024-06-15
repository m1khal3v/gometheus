package api

import (
	"fmt"
	"net/http"
)

func (container Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	for _, metric := range container.storage.GetAll() {
		_, _ = writer.Write([]byte(fmt.Sprintf("%s: %s\n", metric.GetName(), metric.GetStringValue())))
	}
}
