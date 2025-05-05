// Package retry
// contains Retry helper function
package retry

import (
	"time"
)

type RetryOptions struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Attempts   uint64
	Multiplier uint64
}

// Retry - repeats the function execution a specified number of attempts
// with increasing wait time between them. Stops retrying if the error does not pass the filter
func Retry(
	options RetryOptions,
	function func() error,
	filter func(err error) bool,
) error {
	var err error
	for i := uint64(0); i < options.Attempts; i++ {
		if err = function(); err != nil && (filter == nil || filter(err)) {
			time.Sleep(calculateDelay(options, i))

			continue
		}

		return err
	}

	return err
}

func pow(x, y uint64) uint64 {
	if y == 0 {
		return 1
	}

	if y == 1 {
		return x
	}

	result := x
	for i := uint64(2); i <= y; i++ {
		result *= x
	}
	return result
}

func calculateDelay(options RetryOptions, attempt uint64) time.Duration {
	if attempt == 0 {
		return min(options.BaseDelay, options.MaxDelay)
	}

	return min(options.BaseDelay*time.Duration(pow(options.Multiplier, attempt)), options.MaxDelay)
}
