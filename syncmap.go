package main

import "sync"

type SyncMap struct {
	sync.RWMutex
	internal map[string]map[string]int
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		internal: make(map[string]map[string]int),
	}
}

func (rm *SyncMap) Load(keyA, keyB string) (value int, ok bool) {
	rm.RLock()
	result, ok := rm.internal[keyA][keyB]
	if !ok {
		result = 0
		rm.internal[keyA] = make(map[string]int)
	}
	rm.RUnlock()
	return result, ok
}

func (rm *SyncMap) LoadUnsafe(keyA, keyB string) (value int, ok bool) {
	result, ok := rm.internal[keyA][keyB]
	if !ok {
		result = 0
		rm.internal[keyA] = make(map[string]int)
	}
	return result, ok
}

func (rm *SyncMap) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *SyncMap) Store(keyA, keyB string, value int) {
	rm.Lock()
	rm.internal[keyA][keyB] = value
	rm.Unlock()
}

func (rm *SyncMap) IncrementCount(keyA, keyB string) {
	rm.Lock()
	val, ok := rm.internal[keyA][keyB]
	if !ok {
		val = 0
		if rm.internal[keyA] == nil {
			rm.internal[keyA] = make(map[string]int)
		}
	}
	rm.internal[keyA][keyB] = val + 1
	rm.Unlock()
}

func (rm *SyncMap) NumEntries(keyA string) int {
	rm.RLock()
	m, ok := rm.internal[keyA]
	if !ok {
		return 0

	}
	l := len(m)
	rm.RUnlock()
	return l
}

func (rm *SyncMap) IncrementCountUnsafe(keyA, keyB string) {
	val, ok := rm.internal[keyA][keyB]
	if !ok {
		val = 0
		if rm.internal[keyA] == nil {
			rm.internal[keyA] = make(map[string]int)
		}
	}
	rm.internal[keyA][keyB] = val + 1
}
