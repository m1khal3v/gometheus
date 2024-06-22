package mutex

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestNamedMutex_GlobalLock(t *testing.T) {
	namedMutex := NewNamedMutex()
	namedMutex.GlobalLock()
	assert.False(t, namedMutex.TryLock("test"))
	assert.False(t, namedMutex.TryGlobalLock())
	namedMutex.GlobalUnlock()
	assert.True(t, namedMutex.TryLock("test"))
	assert.True(t, namedMutex.TryGlobalLock())
	namedMutex.GlobalUnlock()
	namedMutex.Unlock("test")
}
