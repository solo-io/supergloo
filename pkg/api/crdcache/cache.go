package crdcache

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
)

var (
	cacheMap = sync.Map{}
)

func storeCache(key interface{}, cache kube.SharedCache) {
	cacheMap.Store(key, cache)
}

func loadCache(key interface{}) kube.SharedCache {
	val, ok := cacheMap.Load(key)
	if !ok {
		return nil
	}
	cache, ok := val.(kube.SharedCache)
	if !ok {
		return nil
	}
	return cache
}

// returns cache, created
func GetCrdCache(ctx context.Context, key interface{}) (kube.SharedCache, bool) {
	if val := loadCache(key); val != nil {
		return val, false
	}
	newCache := kube.NewKubeCache(ctx)
	storeCache(key, newCache)
	return newCache, true
}
