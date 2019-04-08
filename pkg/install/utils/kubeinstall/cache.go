package kubeinstall

import (
	"context"
	"sync"

	"github.com/solo-io/supergloo/pkg/install/utils/kuberesource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

/*
Contains a snapshot of all installed resources
Starts with a snapshot of everytihng in cluster
Warning: takes about 30-45s (in testing) to initialize this cache
*/
type Cache struct {
	access    sync.RWMutex
	resources kuberesource.UnstructuredResourcesByKey
}

func NewCache() *Cache {
	return &Cache{}
}

/*
Initialize the cache with the snapshot of the current cluster
*/
func (c *Cache) Init(ctx context.Context, cfg *rest.Config, filterFuncs ...kuberesource.FilterResource) error {
	// lock the cache at the start of the sync, block all access until sync is complete
	c.access.Lock()
	defer c.access.Unlock()
	currentResources, err := kuberesource.GetClusterResources(ctx, cfg, filterFuncs...)
	if err != nil {
		return err
	}
	c.resources = currentResources.ByKey()
	return nil
}

func (c *Cache) List() kuberesource.UnstructuredResources {
	c.access.RLock()
	defer c.access.RUnlock()
	return c.resources.List()
}

func (c *Cache) Get(key kuberesource.ResourceKey) *unstructured.Unstructured {
	c.access.RLock()
	defer c.access.RUnlock()
	return c.resources[key]
}

func (c *Cache) Set(obj *unstructured.Unstructured) {
	c.access.Lock()
	defer c.access.Unlock()
	c.resources[kuberesource.Key(obj)] = obj
}

func (c *Cache) Delete(obj *unstructured.Unstructured) {
	c.access.Lock()
	defer c.access.Unlock()
	delete(c.resources, kuberesource.Key(obj))
}
