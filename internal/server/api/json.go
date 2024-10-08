package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	pkgErrors "github.com/m1khal3v/gometheus/pkg/errors"
	"github.com/m1khal3v/gometheus/pkg/response"
	"go.uber.org/zap"
)

func DecodeAndValidateJSONRequest[T any](request *http.Request, writer http.ResponseWriter) (*T, bool) {
	target := new(T)

	if err := json.NewDecoder(request.Body).Decode(target); err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return nil, false
	}

	if _, err := govalidator.ValidateStruct(target); err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid request received", err)
		return nil, false
	}

	return target, true
}

func DecodeAndValidateJSONRequests[T any](request *http.Request, writer http.ResponseWriter) ([]*T, bool) {
	targets := make([]*T, 0)

	if err := json.NewDecoder(request.Body).Decode(&targets); err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return nil, false
	}

	if len(targets) == 0 {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Empty request received", nil)
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
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid request received", errors.Join(errs...))
		return nil, false
	}

	return targets, true
}

func WriteJSONResponse(response any, writer http.ResponseWriter) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t encode response", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if _, err = writer.Write(jsonResponse); err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t write response", err)
		return
	}
}

func WriteJSONErrorResponse(status int, writer http.ResponseWriter, message string, responseError error) {
	response := response.APIError{
		Code:    status,
		Message: message,
		Details: pkgErrors.ToUniqueStrings(pkgErrors.Unwrap(responseError)...),
	}

	if status >= 500 {
		logger.Logger.Error(message, zap.Any("details", response.Details))
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if _, err = writer.Write(jsonResponse); err != nil {
		logger.Logger.Error("Failed to write response", zap.Error(err))
	}
}
