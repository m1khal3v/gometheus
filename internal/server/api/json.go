package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	pkgErrors "github.com/m1khal3v/gometheus/pkg/errors"
	"github.com/m1khal3v/gometheus/pkg/response"
	"net/http"
	"strings"
)

func decodeAndValidateJsonRequest[T any](request *http.Request, writer http.ResponseWriter) (*T, bool) {
	var target *T

	if err := json.NewDecoder(request.Body).Decode(target); err != nil {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return nil, false
	}

	if _, err := govalidator.ValidateStruct(target); err != nil {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Invalid request received", err)
		return nil, false
	}

	return target, true
}

func decodeAndValidateJsonRequests[T any](request *http.Request, writer http.ResponseWriter) ([]*T, bool) {
	var targets []*T

	if err := json.NewDecoder(request.Body).Decode(&targets); err != nil {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return nil, false
	}

	if len(targets) == 0 {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Empty request received", nil)
		return nil, false
	}

	var errs []error
	for _, target := range targets {
		if _, err := govalidator.ValidateStruct(target); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Invalid request received", errors.Join(errs...))
		return nil, false
	}

	return targets, true
}

func writeJsonResponse(response any, writer http.ResponseWriter) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		writeJsonErrorResponse(http.StatusInternalServerError, writer, "Can`t encode response", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if _, err = writer.Write(jsonResponse); err != nil {
		writeJsonErrorResponse(http.StatusInternalServerError, writer, "Can`t write response", err)
		return
	}
}

func writeJsonErrorResponse(status int, writer http.ResponseWriter, message string, responseError error) {
	response := response.ApiError{
		Code:    status,
		Message: message,
		Details: pkgErrors.ToUniqueStrings(pkgErrors.Unwrap(responseError)...),
	}

	if status >= 500 {
		logger.Logger.Error(fmt.Sprintf("%s: %s", message, strings.Join(response.Details, "; ")))
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if _, err = writer.Write(jsonResponse); err != nil {
		logger.Logger.Error(err.Error())
	}
}
