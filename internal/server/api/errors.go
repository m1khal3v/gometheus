package api

import (
	"encoding/json"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/pkg/response"
	"net/http"
	"slices"
	"strings"
)

func (container Container) writeErrorResponse(status int, writer http.ResponseWriter, message string, responseError error) {
	response := response.ApiError{
		Code:    status,
		Message: message,
		Details: container.errorsToUniqueStrings(container.unwrapErrors(responseError)),
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

func (container Container) unwrapErrors(err error) []error {
	if err == nil {
		return []error{}
	}
	var errs []error

	wrappedError, ok := err.(interface {
		Unwrap() error
	})
	if ok {
		errs = append(errs, err)
		unwrappedErrors := container.unwrapErrors(wrappedError.Unwrap())
		errs = append(errs, unwrappedErrors...)

		return errs
	}

	joinedError, ok := err.(interface {
		Unwrap() []error
	})
	if ok {
		for _, wrappedError := range joinedError.Unwrap() {
			unwrappedErrors := container.unwrapErrors(wrappedError)
			errs = append(errs, unwrappedErrors...)
		}

		return errs
	}

	multipleErrors, ok := err.(interface {
		Errors() []error
	})
	if ok {
		for _, wrappedError := range multipleErrors.Errors() {
			unwrappedErrors := container.unwrapErrors(wrappedError)
			errs = append(errs, unwrappedErrors...)
		}

		return errs
	}

	return []error{err}
}

func (container Container) errorsToUniqueStrings(errs []error) []string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	slices.Sort(messages)
	return slices.Compact(messages)
}
