package kube

import (
	k8s_core_v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

const KubeConfigSecretLabel = "solo.io/kubeconfig"

type KubeConfig struct {
	// the actual kubeconfig
	Config api.Config
	// expected to be used as an identifier string for a cluster
	// stored as the key for the kubeconfig data in a kubernetes secret
	Cluster string
}

// package up all forms of the config in one convenience struct
type ConvertedConfigs struct {
	ClientConfig clientcmd.ClientConfig
	ApiConfig    *api.Config
	RestConfig   *rest.Config
}

type Converter interface {
	// convert a kube config to the properly-formatted secret
	// If the kube config contains a reference to any files, those are read and the bytes moved to the in-memory secret
	ConfigToSecret(secretName string, secretNamespace string, config *KubeConfig) (*k8s_core_v1.Secret, error)

	// parse a secret out into the cluster it corresponds to as well as all the kube config formats you may need
	SecretToConfig(secret *k8s_core_v1.Secret) (clusterName string, config *ConvertedConfigs, err error)
}

type UnstructuredKubeClient interface {
	// build unstructured k8s objects out of the string representation of a manifest
	BuildResources(namespace string, manifest string) ([]*resource.Info, error)

	// create as many resources as we can, returning the ones that were successful along with any error that occurred
	Create(namespace string, resources []*resource.Info) (createdResources []*resource.Info, err error)

	// delete as many of the given resources as we can
	Delete(namespace string, resources []*resource.Info) (deletedResources []*resource.Info, err error)
}

type UnstructuredKubeClientFactory func(restClientGetter resource.RESTClientGetter) UnstructuredKubeClient
