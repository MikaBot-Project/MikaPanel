package util

import "sync"

type SafeMap[T any] struct {
	data map[string]T
	mu   sync.RWMutex
}

func NewSafeMap[T any]() *SafeMap[T] {
	return &SafeMap[T]{
		data: make(map[string]T),
	}
}

// Set 写操作使用写锁
func (sm *SafeMap[T]) Set(key string, value T) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}

// Get 读操作使用读锁
func (sm *SafeMap[T]) Get(key string) (T, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, exists := sm.data[key]
	return value, exists
}

func (sm *SafeMap[T]) GetMap() map[string]T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.data
}

// Delete 删除操作
func (sm *SafeMap[T]) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

// Range 遍历操作
func (sm *SafeMap[T]) Range(f func(key string, value T) bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for k, v := range sm.data {
		if !f(k, v) {
			break
		}
	}
}
