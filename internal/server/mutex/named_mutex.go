package mutex

import "sync"

type NamedMutex struct {
	mutexMap map[string]*sync.Mutex
	mapMutex sync.Mutex
	locked   bool
}

func NewNamedMutex() *NamedMutex {
	return &NamedMutex{mutexMap: make(map[string]*sync.Mutex)}
}

func (namedMutex *NamedMutex) getLock(name string) *sync.Mutex {
	namedMutex.mapMutex.Lock()
	defer namedMutex.mapMutex.Unlock()

	return namedMutex.createOrGetLock(name)
}

func (namedMutex *NamedMutex) tryGetLock(name string) *sync.Mutex {
	if locked := namedMutex.mapMutex.TryLock(); !locked {
		return nil
	}
	defer namedMutex.mapMutex.Unlock()

	return namedMutex.createOrGetLock(name)
}

func (namedMutex *NamedMutex) createOrGetLock(name string) *sync.Mutex {
	mutex, ok := namedMutex.mutexMap[name]
	if !ok {
		mutex = &sync.Mutex{}
		namedMutex.mutexMap[name] = mutex
	}

	return mutex
}

func (namedMutex *NamedMutex) TryLock(name string) bool {
	lock := namedMutex.tryGetLock(name)
	if lock == nil {
		return false
	}

	return lock.TryLock()
}

func (namedMutex *NamedMutex) Lock(name string) {
	namedMutex.getLock(name).Lock()
}

func (namedMutex *NamedMutex) Unlock(name string) {
	namedMutex.getLock(name).Unlock()
}

func (namedMutex *NamedMutex) TryGlobalLock() bool {
	return namedMutex.mapMutex.TryLock()
}

func (namedMutex *NamedMutex) GlobalLock() {
	namedMutex.mapMutex.Lock()
}

func (namedMutex *NamedMutex) GlobalUnlock() {
	namedMutex.mapMutex.Unlock()
}
