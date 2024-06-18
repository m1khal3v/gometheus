package mutex

import "sync"

type NamedMutex struct {
	mutexMap map[string]*sync.Mutex
	mapMutex sync.Mutex
}

func NewNamedMutex() *NamedMutex {
	return &NamedMutex{mutexMap: make(map[string]*sync.Mutex)}
}

func (namedMutex *NamedMutex) getLock(name string) *sync.Mutex {
	namedMutex.mapMutex.Lock()
	defer namedMutex.mapMutex.Unlock()

	mutex, ok := namedMutex.mutexMap[name]
	if !ok {
		mutex = &sync.Mutex{}
		namedMutex.mutexMap[name] = mutex
	}

	return mutex
}

func (namedMutex *NamedMutex) TryLock(name string) bool {
	return namedMutex.getLock(name).TryLock()
}

func (namedMutex *NamedMutex) Lock(name string) {
	namedMutex.getLock(name).Lock()
}

func (namedMutex *NamedMutex) Unlock(name string) {
	namedMutex.getLock(name).Unlock()
}

func (namedMutex *NamedMutex) TryLockAll() bool {
	return namedMutex.mapMutex.TryLock()
}

func (namedMutex *NamedMutex) LockAll() {
	namedMutex.mapMutex.Lock()
}

func (namedMutex *NamedMutex) UnlockAll() {
	namedMutex.mapMutex.Unlock()
}
