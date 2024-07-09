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
			if i == 0 {
				time.Sleep(baseDelay)
			} else {
				time.Sleep(baseDelay * time.Duration(pow(multiplier, i)))
			}

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
