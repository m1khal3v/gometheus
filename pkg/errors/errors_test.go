package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type wrapped struct {
}

func (wrapped *wrapped) Error() string {
	return "wrapped"
}
func (wrapped *wrapped) Unwrap() error {
	return errors.New("unwrapped")
}

type joined struct {
}

func (joined *joined) Error() string {
	return "wrapped"
}
func (joined *joined) Unwrap() []error {
	return []error{
		errors.New("joined 1"),
		errors.New("joined 2"),
	}
}

type multiple struct {
}

func (multiple *multiple) Error() string {
	return "wrapped"
}
func (multiple *multiple) Errors() []error {
	return []error{
		errors.New("multiple 1"),
		errors.New("multiple 2"),
	}
}

func TestUnwrap(t *testing.T) {
	result := Unwrap(errors.Join(
		errors.Join(new(wrapped), new(joined), new(multiple)),
		errors.Join(new(wrapped), new(joined), new(multiple)),
	))
	assert.Equal(t, []error{
		new(wrapped),
		errors.New("unwrapped"),
		errors.New("joined 1"),
		errors.New("joined 2"),
		errors.New("multiple 1"),
		errors.New("multiple 2"),
		new(wrapped),
		errors.New("unwrapped"),
		errors.New("joined 1"),
		errors.New("joined 2"),
		errors.New("multiple 1"),
		errors.New("multiple 2"),
	}, result)
}

func TestToUniqueStrings(t *testing.T) {
	result := ToUniqueStrings(Unwrap(errors.Join(
		errors.Join(new(wrapped), new(joined), new(multiple)),
		errors.Join(new(wrapped), new(joined), new(multiple)),
	))...)
	assert.ElementsMatch(t, []string{
		"wrapped",
		"unwrapped",
		"joined 1",
		"joined 2",
		"multiple 1",
		"multiple 2",
	}, result)
}
