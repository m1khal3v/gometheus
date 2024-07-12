package retry

import (
	"time"
)

func Retry(
	baseDelay time.Duration,
	attempts uint64,
	multiplier uint64,
	function func() error,
	filter func(err error) bool,
) error {
	var err error
	for i := uint64(0); i < attempts; i++ {
		if err = function(); err != nil && filter(err) {
			time.Sleep(calculateDelay(baseDelay, i, multiplier))

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

func calculateDelay(baseDelay time.Duration, attempt uint64, multiplier uint64) time.Duration {
	if attempt == 0 {
		return baseDelay
	} else {
		return baseDelay * time.Duration(pow(multiplier, attempt))
	}
}
