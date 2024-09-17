package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedMutex_Lock(t *testing.T) {
	namedMutex := NewNamedMutex()
	namedMutex.Lock("test")
	assert.False(t, namedMutex.TryLock("test"))
	assert.True(t, namedMutex.TryLock("test2"))
	assert.False(t, namedMutex.TryLock("test2"))
	namedMutex.Unlock("test")
	assert.True(t, namedMutex.TryLock("test"))
	namedMutex.Unlock("test")
	namedMutex.Unlock("test2")
}
