package smi

import (
	"sort"

	"github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	splitclient "github.com/deislabs/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/api/external/smi/split"
	sgsplit "github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type trafficSplitBaseClient struct {
	splitClient splitclient.Interface
	cache       Cache
}

func NewTrafficSplitClient(splitClient splitclient.Interface, cache Cache) sgsplit.TrafficSplitClient {
	return sgsplit.NewTrafficSplitClientWithBase(&trafficSplitBaseClient{
		splitClient: splitClient,
		cache:       cache,
	})
}

func TrafficSplitFromKube(sp *v1alpha1.TrafficSplit) *sgsplit.TrafficSplit {
	deepCopy := sp.DeepCopy()
	baseType := split.TrafficSplit(*deepCopy)
	resource := &sgsplit.TrafficSplit{
		TrafficSplit: baseType,
	}

	return resource
}

func TrafficSplitToKube(resource resources.Resource) (*v1alpha1.TrafficSplit, error) {
	trafficSplitResource, ok := resource.(*sgsplit.TrafficSplit)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to service profile client", resources.Kind(resource))
	}

	trafficSplit := v1alpha1.TrafficSplit(trafficSplitResource.TrafficSplit)

	return &trafficSplit, nil
}

var _ clients.ResourceClient = &trafficSplitBaseClient{}

func (rc *trafficSplitBaseClient) Kind() string {
	return resources.Kind(&sgsplit.TrafficSplit{})
}

func (rc *trafficSplitBaseClient) NewResource() resources.Resource {
	return resources.Clone(&sgsplit.TrafficSplit{})
}

func (rc *trafficSplitBaseClient) Register() error {
	return nil
}

func (rc *trafficSplitBaseClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	trafficSplitObj, err := rc.splitClient.SmispecV1alpha1().TrafficSplits(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading trafficSplitObj from kubernetes")
	}
	resource := TrafficSplitFromKube(trafficSplitObj)

	if resource == nil {
		return nil, errors.Errorf("trafficSplitObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *trafficSplitBaseClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	trafficSplitObj, err := TrafficSplitToKube(resource)
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
		if _, err := rc.splitClient.SmispecV1alpha1().TrafficSplits(meta.Namespace).Update(trafficSplitObj); err != nil {
			return nil, errors.Wrapf(err, "updating kube trafficSplitObj %v", trafficSplitObj.Name)
		}
	} else {
		if _, err := rc.splitClient.SmispecV1alpha1().TrafficSplits(meta.Namespace).Create(trafficSplitObj); err != nil {
			return nil, errors.Wrapf(err, "creating kube trafficSplitObj %v", trafficSplitObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(trafficSplitObj.Namespace, trafficSplitObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *trafficSplitBaseClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.splitClient.SmispecV1alpha1().TrafficSplits(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting trafficSplitObj %v", name)
	}
	return nil
}

func (rc *trafficSplitBaseClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	trafficSplitObjList, err := rc.cache.TrafficSplitLister().TrafficSplits(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing trafficSplits level")
	}
	var resourceList resources.ResourceList
	for _, trafficSplitObj := range trafficSplitObjList {
		resource := TrafficSplitFromKube(trafficSplitObj)

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

func (rc *trafficSplitBaseClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
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

func (rc *trafficSplitBaseClient) exist(namespace, name string) bool {
	_, err := rc.splitClient.SmispecV1alpha1().TrafficSplits(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
