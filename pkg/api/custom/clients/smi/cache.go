package smi

import (
	"context"
	"sync"
	"time"

	accessclient "github.com/deislabs/smi-sdk-go/pkg/gen/client/access/clientset/versioned"
	accessinformer "github.com/deislabs/smi-sdk-go/pkg/gen/client/access/informers/externalversions"
	accessv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/gen/client/access/listers/access/v1alpha1"
	specsclient "github.com/deislabs/smi-sdk-go/pkg/gen/client/specs/clientset/versioned"
	specsinformer "github.com/deislabs/smi-sdk-go/pkg/gen/client/specs/informers/externalversions"
	specsv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/gen/client/specs/listers/specs/v1alpha1"
	splitclient "github.com/deislabs/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	splitinformer "github.com/deislabs/smi-sdk-go/pkg/gen/client/split/informers/externalversions"
	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/gen/client/split/listers/split/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
)

type Cache interface {
	TrafficTargetLister() accessv1alpha1.TrafficTargetLister
	HTTPRouteGroupLister() specsv1alpha1.HTTPRouteGroupLister
	TrafficSplitLister() splitv1alpha1.TrafficSplitLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type smiCache struct {
	TrafficTargets  accessv1alpha1.TrafficTargetLister
	HTTPRouteGroups specsv1alpha1.HTTPRouteGroupLister
	TrafficSplits   splitv1alpha1.TrafficSplitLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewSMICache(ctx context.Context,
	accessClient accessclient.Interface,
	specsClient specsclient.Interface,
	splitClient splitclient.Interface) (*smiCache, error) {
	resyncDuration := 12 * time.Hour

	accessInformerFactory := accessinformer.NewSharedInformerFactory(accessClient, resyncDuration)
	specsInformerFactory := specsinformer.NewSharedInformerFactory(specsClient, resyncDuration)
	splitInformerFactory := splitinformer.NewSharedInformerFactory(splitClient, resyncDuration)

	trafficTargets := accessInformerFactory.Access().V1alpha1().TrafficTargets()
	httpRouteGroups := specsInformerFactory.Specs().V1alpha1().HTTPRouteGroups()
	trafficSplits := splitInformerFactory.Smispec().V1alpha1().TrafficSplits()

	k := &smiCache{
		TrafficTargets:  trafficTargets.Lister(),
		HTTPRouteGroups: httpRouteGroups.Lister(),
		TrafficSplits:   trafficSplits.Lister(),
	}

	kubeController := controller.NewController("linkerd-resources-cache",
		controller.NewLockingSyncHandler(k.updatedOccured),
		trafficTargets.Informer(),
		httpRouteGroups.Informer(),
		trafficSplits.Informer(),
	)

	stop := ctx.Done()
	err := kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *smiCache) TrafficTargetLister() accessv1alpha1.TrafficTargetLister {
	return k.TrafficTargets
}

func (k *smiCache) HTTPRouteGroupLister() specsv1alpha1.HTTPRouteGroupLister {
	return k.HTTPRouteGroups
}

func (k *smiCache) TrafficSplitLister() splitv1alpha1.TrafficSplitLister {
	return k.TrafficSplits
}

func (k *smiCache) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *smiCache) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *smiCache) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
