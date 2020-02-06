package concurrency

import "sync"

// interface around the sync.Map type so that we can mock it
//go:generate mockgen -source threadsafe_map.go -destination mocks/mock_threadsafe_map.go
type ThreadSafeMap interface {
	Delete(key interface{})
	Load(key interface{}) (value interface{}, ok bool)
	LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
	Range(f func(key, value interface{}) bool)
	Store(key, value interface{})
}

func NewThreadSafeMap() ThreadSafeMap {
	return &threadSafeMap{}
}

type threadSafeMap struct {
	syncMap sync.Map
}

func (t *threadSafeMap) Delete(key interface{}) {
	t.syncMap.Delete(key)
}

func (t *threadSafeMap) Load(key interface{}) (value interface{}, ok bool) {
	return t.syncMap.Load(key)
}

func (t *threadSafeMap) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	return t.syncMap.LoadOrStore(key, value)
}

func (t *threadSafeMap) Range(f func(key, value interface{}) bool) {
	t.syncMap.Range(f)
}

func (t *threadSafeMap) Store(key, value interface{}) {
	t.syncMap.Store(key, value)
}
