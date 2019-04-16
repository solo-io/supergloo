package linkerd

import (
	"context"
	"sync"
	"time"

	linkerdclient "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	linkerdinformers "github.com/linkerd/linkerd2/controller/gen/client/informers/externalversions"
	"github.com/linkerd/linkerd2/controller/gen/client/listers/serviceprofile/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
)

type Cache interface {
	ServiceProfileLister() v1alpha1.ServiceProfileLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type linkerdCache struct {
	serviceProfiles v1alpha1.ServiceProfileLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewLinkerdCache(ctx context.Context, linkerdClient linkerdclient.Interface) (*linkerdCache, error) {
	resyncDuration := 12 * time.Hour
	sharedInformerFactory := linkerdinformers.NewSharedInformerFactory(linkerdClient, resyncDuration)

	serviceProfiles := sharedInformerFactory.Linkerd().V1alpha1().ServiceProfiles()

	k := &linkerdCache{
		serviceProfiles: serviceProfiles.Lister(),
	}

	kubeController := controller.NewController("linkerd-resources-cache",
		controller.NewLockingSyncHandler(k.updatedOccured),
		serviceProfiles.Informer())

	stop := ctx.Done()
	err := kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *linkerdCache) ServiceProfileLister() v1alpha1.ServiceProfileLister {
	return k.serviceProfiles
}

func (k *linkerdCache) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *linkerdCache) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *linkerdCache) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
