// Package errors
// contains helper functions for error type
package errors

import (
	"slices"
)

func Unwrap(err error) []error {
	if err == nil {
		return []error{}
	}
	var errs []error

	wrappedError, ok := err.(interface {
		Unwrap() error
	})
	if ok {
		errs = append(errs, err)
		unwrappedErrors := Unwrap(wrappedError.Unwrap())
		errs = append(errs, unwrappedErrors...)

		return errs
	}

	joinedError, ok := err.(interface {
		Unwrap() []error
	})
	if ok {
		for _, wrappedError := range joinedError.Unwrap() {
			unwrappedErrors := Unwrap(wrappedError)
			errs = append(errs, unwrappedErrors...)
		}

		return errs
	}

	multipleErrors, ok := err.(interface {
		Errors() []error
	})
	if ok {
		for _, wrappedError := range multipleErrors.Errors() {
			unwrappedErrors := Unwrap(wrappedError)
			errs = append(errs, unwrappedErrors...)
		}

		return errs
	}

	return []error{err}
}

func ToUniqueStrings(errs ...error) []string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	slices.Sort(messages)
	return slices.Compact(messages)
}
