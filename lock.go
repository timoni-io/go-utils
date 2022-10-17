package utils

import "sync"

type Lock struct {
	lock sync.RWMutex
}

func (l *Lock) Lock() {
	if l != nil {
		l.lock.Lock()
	}
}

func (l *Lock) Unlock() {
	if l != nil {
		l.lock.Unlock()
	}
}

func (l *Lock) RLock() {
	if l != nil {
		l.lock.RLock()
	}
}

func (l *Lock) RUnlock() {
	if l != nil {
		l.lock.RUnlock()
	}
}
