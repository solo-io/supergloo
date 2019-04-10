package kubernetes

import (
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/supergloo/api/custom/kubepod"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type ResourceClient struct {
	kube  kubernetes.Interface
	cache cache.KubeCoreCache
}

func NewResourceClient(kube kubernetes.Interface, cache cache.KubeCoreCache) *ResourceClient {
	return &ResourceClient{
		kube:  kube,
		cache: cache,
	}
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(&v1.Pod{})
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(&v1.Pod{})
}

func (rc *ResourceClient) Register() error {
	return nil
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	podObj, err := rc.kube.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading podObj from kubernetes")
	}
	resource, err := kubepod.FromKube(podObj)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, errors.Errorf("podObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	podObj, err := kubepod.ToKube(resource)
	if err != nil {
		return nil, err
	}

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{
		Ctx: opts.Ctx,
	})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		if meta.ResourceVersion != original.GetMetadata().ResourceVersion {
			return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, meta.ResourceVersion, original.GetMetadata().ResourceVersion)
		}
		if _, err := rc.kube.CoreV1().Pods(meta.Namespace).Update(podObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube podObj %v", podObj.Name)
		}
	} else {
		if _, err := rc.kube.CoreV1().Pods(meta.Namespace).Create(podObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube podObj %v", podObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(podObj.Namespace, podObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.kube.CoreV1().Pods(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting podObj %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	podObjList, err := rc.cache.PodLister().Pods(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing pods level")
	}
	var resourceList resources.ResourceList
	for _, podObj := range podObjList {
		resource, err := kubepod.FromKube(podObj)
		if err != nil {
			return nil, err
		}
		if resource == nil {
			continue
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	watch := rc.cache.Subscribe()

	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	// prevent flooding the channel with duplicates
	var previous *resources.ResourceList
	updateResourceList := func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		if previous != nil {
			if list.Equal(*previous) {
				return
			}
		}
		previous = &list
		resourcesChan <- list
	}

	go func() {
		defer rc.cache.Unsubscribe(watch)
		defer close(resourcesChan)
		defer close(errs)

		// watch should open up with an initial read
		updateResourceList()
		for {
			select {
			case _, ok := <-watch:
				if !ok {
					return
				}
				updateResourceList()
			case <-opts.Ctx.Done():
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

func (rc *ResourceClient) exist(namespace, name string) bool {
	_, err := rc.kube.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
