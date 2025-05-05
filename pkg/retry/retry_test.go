package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_pow(t *testing.T) {
	tests := []struct {
		name string
		x    uint64
		y    uint64
		want uint64
	}{
		{
			name: "2 pow 2",
			x:    2,
			y:    2,
			want: 4,
		},
		{
			name: "10 pow 10",
			x:    10,
			y:    10,
			want: 10000000000,
		},
		{
			name: "3 pow 17",
			x:    3,
			y:    17,
			want: 129140163,
		},
		{
			name: "17 pow 3",
			x:    17,
			y:    3,
			want: 4913,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, pow(tt.x, tt.y))
		})
	}
}

func Test_calculateDelay(t *testing.T) {
	tests := []struct {
		name       string
		baseDelay  time.Duration
		maxDelay   time.Duration
		attempt    uint64
		multiplier uint64
		want       time.Duration
	}{
		{
			name:       "1/1",
			baseDelay:  time.Second,
			maxDelay:   10 * time.Second,
			attempt:    0,
			multiplier: 2,
			want:       time.Second,
		},
		{
			name:       "1/2",
			baseDelay:  time.Second,
			maxDelay:   10 * time.Second,
			attempt:    1,
			multiplier: 2,
			want:       2 * time.Second,
		},
		{
			name:       "1/3",
			baseDelay:  time.Second,
			maxDelay:   10 * time.Second,
			attempt:    2,
			multiplier: 2,
			want:       4 * time.Second,
		},
		{
			name:       "1/4",
			baseDelay:  time.Second,
			maxDelay:   10 * time.Second,
			attempt:    3,
			multiplier: 2,
			want:       8 * time.Second,
		},
		{
			name:       "2/1",
			baseDelay:  2 * time.Second,
			maxDelay:   15 * time.Second,
			attempt:    0,
			multiplier: 3,
			want:       2 * time.Second,
		},
		{
			name:       "2/2",
			baseDelay:  2 * time.Second,
			maxDelay:   15 * time.Second,
			attempt:    1,
			multiplier: 3,
			want:       6 * time.Second,
		},
		{
			name:       "2/3",
			baseDelay:  2 * time.Second,
			maxDelay:   15 * time.Second,
			attempt:    2,
			multiplier: 3,
			want:       15 * time.Second,
		},
		{
			name:       "2/4",
			baseDelay:  2 * time.Second,
			maxDelay:   15 * time.Second,
			attempt:    3,
			multiplier: 3,
			want:       15 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateDelay(RetryOptions{BaseDelay: tt.baseDelay, MaxDelay: tt.maxDelay, Multiplier: tt.multiplier}, tt.attempt))
		})
	}
}

func TestRetry(t *testing.T) {
	attempts := uint64(0)
	tests := []struct {
		name         string
		attempts     uint64
		wantAttempts uint64
		function     func() error
		filter       func(err error) bool
		wantErr      bool
	}{
		{
			name:         "first attempt ok",
			attempts:     1,
			wantAttempts: 1,
			function: func() error {
				return nil
			},
			filter: func(err error) bool {
				return true
			},
			wantErr: false,
		},
		{
			name:         "non-retryable error",
			attempts:     3,
			wantAttempts: 1,
			function: func() error {
				return errors.New("test error")
			},
			filter: func(err error) bool {
				return err.Error() != "test error"
			},
			wantErr: true,
		},
		{
			name:         "retryable error fail",
			attempts:     3,
			wantAttempts: 3,
			function: func() error {
				return errors.New("test error")
			},
			filter: func(err error) bool {
				return err.Error() == "test error"
			},
			wantErr: true,
		},
		{
			name:         "retryable error ok",
			attempts:     3,
			wantAttempts: 2,
			function: func() error {
				if attempts > 1 {
					return nil
				}
				return errors.New("test error")
			},
			filter: func(err error) bool {
				return err.Error() == "test error"
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			function := func() error {
				attempts++
				return tt.function()
			}
			err := Retry(RetryOptions{
				BaseDelay:  0,
				MaxDelay:   0,
				Attempts:   tt.attempts,
				Multiplier: 0,
			}, function, tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.wantAttempts, attempts)
			attempts = 0
		})
	}
}

func Test_pow_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		x    uint64
		y    uint64
		want uint64
	}{
		{
			name: "0 pow 0",
			x:    0,
			y:    0,
			want: 1, // Стандартное определение
		},
		{
			name: "0 pow 5",
			x:    0,
			y:    5,
			want: 0,
		},
		{
			name: "5 pow 0",
			x:    5,
			y:    0,
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, pow(tt.x, tt.y))
		})
	}
}

func Test_calculateDelay_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		baseDelay  time.Duration
		maxDelay   time.Duration
		attempt    uint64
		multiplier uint64
		want       time.Duration
	}{
		{
			name:       "maxDelay less than baseDelay",
			baseDelay:  2 * time.Second,
			maxDelay:   1 * time.Second,
			attempt:    2,
			multiplier: 2,
			want:       1 * time.Second,
		},
		{
			name:       "multiplier is 1",
			baseDelay:  time.Second,
			maxDelay:   10 * time.Second,
			attempt:    3,
			multiplier: 1,
			want:       time.Second, // Задержка не меняется
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateDelay(RetryOptions{BaseDelay: tt.baseDelay, MaxDelay: tt.maxDelay, Multiplier: tt.multiplier}, tt.attempt))
		})
	}
}

func TestRetry_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		attempts     uint64
		wantAttempts uint64
		function     func() error
		filter       func(err error) bool
		wantErr      bool
	}{
		{
			name:         "nil filter allows all retries",
			attempts:     3,
			wantAttempts: 3,
			function: func() error {
				return errors.New("some error")
			},
			filter:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		attempts := uint64(0)
		t.Run(tt.name, func(t *testing.T) {
			function := func() error {
				attempts++
				return tt.function()
			}
			err := Retry(RetryOptions{
				BaseDelay:  0,
				MaxDelay:   0,
				Attempts:   tt.attempts,
				Multiplier: 0,
			}, function, tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantAttempts, attempts)
		})
	}
}
