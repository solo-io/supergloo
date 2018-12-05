package kube

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	MCreates        = stats.Int64("kube/creates", "The number of creates", "1")
	CreateCountView = &view.View{
		Name:        "kube/creates-count",
		Measure:     MCreates,
		Description: "The number of list calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
		},
	}
	MUpdates        = stats.Int64("kube/updates", "The number of updates", "1")
	UpdateCountView = &view.View{
		Name:        "kube/updates-count",
		Measure:     MUpdates,
		Description: "The number of list calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
		},
	}

	MDeletes        = stats.Int64("kube/deletes", "The number of deletes", "1")
	DeleteCountView = &view.View{
		Name:        "kube/deletes-count",
		Measure:     MDeletes,
		Description: "The number of list calls",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{
			KeyKind,
		},
	}

	KeyOpKind, _ = tag.NewKey("op")

	MInFlight       = stats.Int64("kube/req_in_flight", "The number of requests in flight", "1")
	InFlightSumView = &view.View{
		Name:        "kube/req-in-flight",
		Measure:     MInFlight,
		Description: "The number of list calls",
		Aggregation: view.Sum(),
		TagKeys: []tag.Key{
			KeyOpKind,
			KeyKind,
		},
	}

	MEvents         = stats.Int64("kube/events", "The number of events", "1")
	EventsCountView = &view.View{
		Name:        "kube/events-count",
		Measure:     MEvents,
		Description: "The number of events sent from kuberenets to us",
		Aggregation: view.Count(),
	}
)

func init() {
	view.Register(CreateCountView, UpdateCountView, DeleteCountView, InFlightSumView, EventsCountView)
}

type KubeCache struct {
	sharedInformerFactory     *ResourceClientSharedInformerFactory
	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex

	factoryStarter sync.Once
}

func NewKubeCache() *KubeCache {
	return &KubeCache{
		sharedInformerFactory: NewResourceClientSharedInformerFactory(),
	}
}

func (kc *KubeCache) addWatch() <-chan struct{} {
	kc.cacheUpdatedWatchersMutex.Lock()
	defer kc.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 1)
	kc.cacheUpdatedWatchers = append(kc.cacheUpdatedWatchers, c)
	return c
}

func (kc *KubeCache) removeWatch(c <-chan struct{}) {
	kc.cacheUpdatedWatchersMutex.Lock()
	defer kc.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range kc.cacheUpdatedWatchers {
		if cacheUpdated == c {
			kc.cacheUpdatedWatchers = append(kc.cacheUpdatedWatchers[:i], kc.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (kc *KubeCache) startFactory(ctx context.Context, client kubernetes.Interface) {
	kc.factoryStarter.Do(func() {
		kc.sharedInformerFactory.Start(ctx, client, kc.updatedOccured)
	})

	// we want to panic here because the initial bootstrap of the cache failed
	// this should be a rare error, and if we are restarted should not happen again
	if err := kc.sharedInformerFactory.InitErr(); err != nil {
		contextutils.LoggerFrom(ctx).Panicf("failed to intiialize kube shared informer factory: %v", err)
	}
}

// TODO(yuval-k): See if we can get more fine grained updates here, about which resources was udpated
func (kc *KubeCache) updatedOccured() {
	stats.Record(context.TODO(), MEvents.M(1))
	kc.cacheUpdatedWatchersMutex.Lock()
	defer kc.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range kc.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}

// lazy start in list & watch
// register informers in register

type ResourceClient struct {
	crd          crd.Crd
	apiexts      apiexts.Interface
	kube         versioned.Interface
	kubeClient   kubernetes.Interface
	ownerLabel   string
	resourceName string
	resourceType resources.InputResource
	sharedCache  *KubeCache
}

func NewResourceClient(crd crd.Crd, cfg *rest.Config, sharedCache *KubeCache, resourceType resources.InputResource) (*ResourceClient, error) {
	apiExts, err := apiexts.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "creating api extensions client")
	}
	crdClient, err := versioned.NewForConfig(cfg, crd)
	if err != nil {
		return nil, errors.Wrapf(err, "creating crd client")
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube clientset: %v", err)
	}

	typeof := reflect.TypeOf(resourceType)
	resourceName := strings.Replace(typeof.String(), "*", "", -1)
	resourceName = strings.Replace(resourceName, ".", "", -1)

	return &ResourceClient{
		crd:          crd,
		apiexts:      apiExts,
		kube:         crdClient,
		kubeClient:   kubeClient,
		resourceName: resourceName,
		resourceType: resourceType,
		sharedCache:  sharedCache,
	}, nil
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(rc.resourceType)
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.resourceType)
}

func (rc *ResourceClient) Register() error {
	/*
		shared informer factory for all namespaces; and then filter desired namespace in list and watch!
		zbam!

	*/
	rc.sharedCache.sharedInformerFactory.Register(rc)
	return rc.crd.Register(rc.apiexts)
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()
	namespace = clients.DefaultNamespaceIfEmpty(namespace)
	ctx := opts.Ctx

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "read")); err == nil {
		ctx = ctxWithTags
	}

	stats.Record(ctx, MInFlight.M(1))
	resourceCrd, err := rc.kube.ResourcesV1().Resources(namespace).Get(name, metav1.GetOptions{})
	stats.Record(ctx, MInFlight.M(-1))
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading resource from kubernetes")
	}
	resource, err := rc.convertCrdToResource(resourceCrd)
	if err != nil {
		return nil, errors.Wrapf(err, "converting output crd")
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()
	meta.Namespace = clients.DefaultNamespaceIfEmpty(meta.Namespace)

	// mutate and return clone
	clone := proto.Clone(resource).(resources.InputResource)
	clone.SetMetadata(meta)
	resourceCrd := rc.crd.KubeResource(clone)

	ctx := opts.Ctx
	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "write")); err == nil {
		ctx = ctxWithTags
	}

	if rc.exist(ctx, meta.Namespace, meta.Name) {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		stats.Record(ctx, MUpdates.M(1), MInFlight.M(1))
		defer stats.Record(ctx, MInFlight.M(-1))
		if _, updateerr := rc.kube.ResourcesV1().Resources(meta.Namespace).Update(resourceCrd); updateerr != nil {
			original, err := rc.kube.ResourcesV1().Resources(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			if err == nil {
				return nil, errors.Wrapf(updateerr, "updating kube resource %v:%v (want %v)", resourceCrd.Name, resourceCrd.ResourceVersion, original.ResourceVersion)
			}
			return nil, errors.Wrapf(updateerr, "updating kube resource %v", resourceCrd.Name)
		}
	} else {
		stats.Record(ctx, MCreates.M(1), MInFlight.M(1))
		defer stats.Record(ctx, MInFlight.M(-1))
		if _, err := rc.kube.ResourcesV1().Resources(meta.Namespace).Create(resourceCrd); err != nil {
			if apierrors.IsAlreadyExists(err) {
				return nil, errors.NewExistErr(meta)
			}
			return nil, errors.Wrapf(err, "creating kube resource %v", resourceCrd.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	ctx := opts.Ctx

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Insert(KeyOpKind, "delete")); err == nil {
		ctx = ctxWithTags
	}
	stats.Record(ctx, MDeletes.M(1))

	if !rc.exist(ctx, namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr(namespace, name)
		}
		return nil
	}

	stats.Record(ctx, MInFlight.M(1))
	defer stats.Record(ctx, MInFlight.M(-1))
	if err := rc.kube.ResourcesV1().Resources(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting resource %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	rc.sharedCache.startFactory(context.TODO(), rc.kubeClient)

	lister, err := rc.sharedCache.sharedInformerFactory.GetLister(rc.crd.Type)
	if err != nil {
		return nil, err
	}
	allResources, err := lister.List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing resources in %v", namespace)
	}
	var listedResources []*v1.Resource
	if namespace != "" {
		for _, r := range allResources {
			if r.ObjectMeta.Namespace == namespace {
				listedResources = append(listedResources, r)
			}
		}
	} else {
		listedResources = allResources
	}

	var resourceList resources.ResourceList
	for _, resourceCrd := range listedResources {
		resource, err := rc.convertCrdToResource(resourceCrd)
		if err != nil {
			return nil, errors.Wrapf(err, "converting output crd")
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	rc.sharedCache.startFactory(context.TODO(), rc.kubeClient)

	opts = opts.WithDefaults()
	namespace = clients.DefaultNamespaceIfEmpty(namespace)
	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	ctx := opts.Ctx

	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		resourcesChan <- list
	}
	// watch should open up with an initial read
	cacheUpdated := rc.sharedCache.addWatch()

	go func() {
		defer rc.sharedCache.removeWatch(cacheUpdated)
		defer close(resourcesChan)
		defer close(errs)

		updateResourceList()

		for {
			select {
			case <-time.After(opts.RefreshRate): // TODO(yuval-k): can we remove this? is the factory takes care of that...
				updateResourceList()
			case <-cacheUpdated:
				updateResourceList()
			case <-ctx.Done():
				return
			}
		}

	}()

	return resourcesChan, errs, nil
}

func (rc *ResourceClient) exist(ctx context.Context, namespace, name string) bool {

	if ctxWithTags, err := tag.New(ctx, tag.Insert(KeyKind, rc.resourceName), tag.Upsert(KeyOpKind, "get")); err == nil {
		ctx = ctxWithTags
	}

	stats.Record(ctx, MInFlight.M(1))
	defer stats.Record(ctx, MInFlight.M(-1))

	_, err := rc.kube.ResourcesV1().Resources(namespace).Get(name, metav1.GetOptions{}) // TODO(yuval-k): check error for real
	return err == nil

}

func (rc *ResourceClient) convertCrdToResource(resourceCrd *v1.Resource) (resources.Resource, error) {
	resource := rc.NewResource()
	if resourceCrd.Spec != nil {
		if err := protoutils.UnmarshalMap(*resourceCrd.Spec, resource); err != nil {
			return nil, errors.Wrapf(err, "reading crd spec into %v", rc.resourceName)
		}
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(resourceCrd.ObjectMeta))
	if withStatus, ok := resource.(resources.InputResource); ok {
		resources.UpdateStatus(withStatus, func(status *core.Status) {
			*status = resourceCrd.Status
		})
	}
	return resource, nil
}
