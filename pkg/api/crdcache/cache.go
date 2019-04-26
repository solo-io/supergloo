package crdcache

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
)

var (
	sharedCache     kube.SharedCache
	sharedCacheLock sync.Mutex
)

func GetCache(ctx context.Context) kube.SharedCache {
	// Lock this function as it can be called by multiple go routines simultaneously
	sharedCacheLock.Lock()
	defer sharedCacheLock.Unlock()
	if sharedCache == nil {
		sharedCache = kube.NewKubeCache(ctx)
	}
	return sharedCache
}

func ResetCache(ctx context.Context) kube.SharedCache {
	sharedCacheLock.Lock()
	defer sharedCacheLock.Unlock()
	sharedCache = kube.NewKubeCache(ctx)
	return sharedCache
}
