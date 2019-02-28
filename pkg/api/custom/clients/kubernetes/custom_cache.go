package kubernetes

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"

	kubeinformers "k8s.io/client-go/informers"
	kubelisters "k8s.io/client-go/listers/core/v1"

	"k8s.io/client-go/kubernetes"
)

type CustomKubeCache interface {
	PodLister() kubelisters.PodLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type CustomKubeCaches struct {
	initError error

	podLister kubelisters.PodLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewCustomKubeCache(ctx context.Context, client kubernetes.Interface) *CustomKubeCaches {
	resyncDuration := 12 * time.Hour
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(client, resyncDuration)

	pods := kubeInformerFactory.Core().V1().Pods()
	k := &CustomKubeCaches{
		podLister: pods.Lister(),
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured),
		pods.Informer())

	stop := ctx.Done()
	go kubeInformerFactory.Start(stop)
	go func() {
		err := kubeController.Run(2, stop)
		if err != nil {
			k.initError = err
		}
	}()

	return k
}

func (k *CustomKubeCaches) PodLister() kubelisters.PodLister {
	return k.podLister
}

func (k *CustomKubeCaches) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *CustomKubeCaches) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *CustomKubeCaches) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
