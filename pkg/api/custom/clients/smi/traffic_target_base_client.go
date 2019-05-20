package smi

import (
	"sort"

	"github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/access/v1alpha1"
	accessclient "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/gen/client/access/clientset/versioned"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/api/external/smi/access"
	sgaccess "github.com/solo-io/supergloo/pkg/api/external/smi/access/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type trafficTargetBaseClient struct {
	accessClient accessclient.Interface
	cache        Cache
}

func NewTrafficTargetClient(accessClient accessclient.Interface, cache Cache) sgaccess.TrafficTargetClient {
	return sgaccess.NewTrafficTargetClientWithBase(&trafficTargetBaseClient{
		accessClient: accessClient,
		cache:        cache,
	})
}

func TrafficTargetFromKube(sp *v1alpha1.TrafficTarget) *sgaccess.TrafficTarget {
	deepCopy := sp.DeepCopy()
	baseType := access.TrafficTarget(*deepCopy)
	resource := &sgaccess.TrafficTarget{
		TrafficTarget: baseType,
	}

	return resource
}

func TrafficTargetToKube(resource resources.Resource) (*v1alpha1.TrafficTarget, error) {
	trafficTargetResource, ok := resource.(*sgaccess.TrafficTarget)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to service profile client", resources.Kind(resource))
	}

	trafficTarget := v1alpha1.TrafficTarget(trafficTargetResource.TrafficTarget)

	return &trafficTarget, nil
}

var _ clients.ResourceClient = &trafficTargetBaseClient{}

func (rc *trafficTargetBaseClient) Kind() string {
	return resources.Kind(&sgaccess.TrafficTarget{})
}

func (rc *trafficTargetBaseClient) NewResource() resources.Resource {
	return resources.Clone(&sgaccess.TrafficTarget{})
}

func (rc *trafficTargetBaseClient) Register() error {
	return nil
}

func (rc *trafficTargetBaseClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	trafficTargetObj, err := rc.accessClient.AccessV1alpha1().TrafficTargets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading trafficTargetObj from kubernetes")
	}
	resource := TrafficTargetFromKube(trafficTargetObj)

	if resource == nil {
		return nil, errors.Errorf("trafficTargetObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *trafficTargetBaseClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	trafficTargetObj, err := TrafficTargetToKube(resource)
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
		if _, err := rc.accessClient.AccessV1alpha1().TrafficTargets(meta.Namespace).Update(trafficTargetObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube trafficTargetObj %v", trafficTargetObj.Name)
		}
	} else {
		if _, err := rc.accessClient.AccessV1alpha1().TrafficTargets(meta.Namespace).Create(trafficTargetObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube trafficTargetObj %v", trafficTargetObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(trafficTargetObj.Namespace, trafficTargetObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *trafficTargetBaseClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.accessClient.AccessV1alpha1().TrafficTargets(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting trafficTargetObj %v", name)
	}
	return nil
}

func (rc *trafficTargetBaseClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	trafficTargetObjList, err := rc.cache.TrafficTargetLister().TrafficTargets(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing trafficTargets level")
	}
	var resourceList resources.ResourceList
	for _, trafficTargetObj := range trafficTargetObjList {
		resource := TrafficTargetFromKube(trafficTargetObj)

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

func (rc *trafficTargetBaseClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
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

func (rc *trafficTargetBaseClient) exist(namespace, name string) bool {
	_, err := rc.accessClient.AccessV1alpha1().TrafficTargets(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
