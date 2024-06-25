package mutex

import (
	"sync"
	"time"
)

const ttl = 5 * time.Minute
const deleteFrequency = 60 * time.Second

type mutexItem struct {
	mutex      *sync.Mutex
	lastAccess time.Time
}

type NamedMutex struct {
	mutexMap *sync.Map
}

func NewNamedMutex() *NamedMutex {
	namedMutex := &NamedMutex{mutexMap: &sync.Map{}}

	go func() {
		ticker := time.NewTicker(deleteFrequency)
		now := time.Now()

		for range ticker.C {
			namedMutex.mutexMap.Range(func(key, value any) bool {
				if now.Sub(value.(*mutexItem).lastAccess) > ttl {
					namedMutex.mutexMap.Delete(key)
				}
				return true
			})
		}
	}()

	return namedMutex
}

func (namedMutex *NamedMutex) createOrGetLock(name string) *sync.Mutex {
	now := time.Now()
	actual, exists := namedMutex.mutexMap.LoadOrStore(name, &mutexItem{
		mutex:      &sync.Mutex{},
		lastAccess: now,
	})

	item := actual.(*mutexItem)
	if exists {
		item.lastAccess = now
	}

	return item.mutex
}

func (namedMutex *NamedMutex) TryLock(name string) bool {
	return namedMutex.createOrGetLock(name).TryLock()
}

func (namedMutex *NamedMutex) Lock(name string) {
	namedMutex.createOrGetLock(name).Lock()
}

func (namedMutex *NamedMutex) Unlock(name string) {
	namedMutex.createOrGetLock(name).Unlock()
}
