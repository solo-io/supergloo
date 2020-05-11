package kube

import (
	"k8s.io/cli-runtime/pkg/resource"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type UnstructuredKubeClient interface {
	// build unstructured k8s objects out of the string representation of a manifest
	BuildResources(namespace string, manifest string) ([]*resource.Info, error)

	// create as many resources as we can, returning the ones that were successful along with any error that occurred
	Create(namespace string, resources []*resource.Info) (createdResources []*resource.Info, err error)

	// delete as many of the given resources as we can
	Delete(namespace string, resources []*resource.Info) (deletedResources []*resource.Info, err error)
}

type UnstructuredKubeClientFactory func(restClientGetter resource.RESTClientGetter) UnstructuredKubeClient
