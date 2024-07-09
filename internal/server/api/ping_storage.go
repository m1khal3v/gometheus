package api

import (
	"net/http"
)

func (container Container) PingStorage(writer http.ResponseWriter, request *http.Request) {
	if err := container.manager.PingStorage(request.Context()); err == nil {
		writer.WriteHeader(http.StatusOK)
	} else {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Storage unavailable", err)
	}
}
