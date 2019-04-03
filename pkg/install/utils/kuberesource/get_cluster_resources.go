package kuberesource

import (
	"context"

	"github.com/goph/emperror"
	"github.com/solo-io/go-utils/contextutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type FilterResource func(schema.GroupVersionResource) bool

/*
Use to get all CRUD'able resources for a cluster
Returns a list of UnstructuredResources which are a wrapper for the
map[string]interface{} type created from Kubernetes-style JSON objects.

Warning: slow function, be sure to call asynchronously.
Filter funcs can be passed to reduce the latency of this function,
i.e.: by excluding resource types (each resource type gets its own
query, contributing to latency of this function).
*/
func GetClusterResources(ctx context.Context, cfg *rest.Config, filterFuncs ...FilterResource) (UnstructuredResources, error) {
	// discovery client
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// resource client
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// list api resources that can be CRUD'ed
	serverResources, err := disc.ServerResources()
	if err != nil {
		return nil, err
	}
	crudableResources := discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"create", "list", "watch", "delete"}}, serverResources)

	gv, err := discovery.GroupVersionResources(crudableResources)
	if err != nil {
		return nil, err
	}
	// convert map to slice
	var groupVersionResources []schema.GroupVersionResource
	for res := range gv {
		groupVersionResources = append(groupVersionResources, res)
	}
	var allResources UnstructuredResources
	for _, gvr := range filterGroupVersions(groupVersionResources, filterFuncs...) {
		gvr := gvr
		contextutils.LoggerFrom(ctx).Infow("listing all", "resourceType", gvr)
		resources, err := client.Resource(gvr).List(metav1.ListOptions{})
		if err != nil {
			return nil, emperror.With(err, "group_version_resource", gvr)
		}

		for i := range resources.Items {
			res := &resources.Items[i]
			contextutils.LoggerFrom(ctx).Infof("appending %v: %v.%v", res.GroupVersionKind().Kind, res.GetNamespace(), res.GetName())
			allResources = append(allResources, res)
		}
	}

	return allResources.Sort(), nil
}

func filterGroupVersions(groupVersions []schema.GroupVersionResource, filterFuncs ...FilterResource) []schema.GroupVersionResource {
	var filteredGroupVersions []schema.GroupVersionResource
	for _, resourceType := range groupVersions {
		var filtered bool
		for _, filter := range filterFuncs {
			if filter(resourceType) {
				filtered = true
				break
			}
		}
		if filtered {
			continue
		}
		filteredGroupVersions = append(filteredGroupVersions, resourceType)
	}
	return filteredGroupVersions
}
