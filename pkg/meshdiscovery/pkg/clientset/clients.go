package clientset

import (
	"context"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	customkube "github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Clientset struct {
	RestConfig *rest.Config

	Kube kubernetes.Interface

	Input *inputClients

	Discovery *discoveryClients
}

func newClientset(restConfig *rest.Config, kube kubernetes.Interface, input *inputClients, discovery *discoveryClients) *Clientset {
	return &Clientset{RestConfig: restConfig, Kube: kube, Input: input, Discovery: discovery}
}

type discoveryClients struct {
	Mesh v1.MeshClient
}

func newDiscoveryClients(mesh v1.MeshClient) *discoveryClients {
	return &discoveryClients{Mesh: mesh}
}

type inputClients struct {
	Pod v1.PodClient
}

func newInputClients(pod v1.PodClient) *inputClients {
	return &inputClients{Pod: pod}
}

func clientForCrd(crd crd.Crd, restConfig *rest.Config, kubeCache kube.SharedCache) factory.ResourceClientFactory {
	return &factory.KubeResourceClientFactory{Crd: crd, Cfg: restConfig, SharedCache: kubeCache}
}

// initialize all resource clients here that will share a cache
func ClientsetFromContext(ctx context.Context) (*Clientset, error) {
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	crdCache := kube.NewKubeCache(ctx)
	kubeCoreCache, err := cache.NewKubeCoreCache(ctx, kubeClient)
	if err != nil {
		return nil, err
	}
	/*
		supergloo config clients
	*/

	mesh, err := v1.NewMeshClient(clientForCrd(v1.MeshCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := mesh.Register(); err != nil {
		return nil, err
	}

	// special resource client wired up to kubernetes pods
	// used by the istio policy syncer to watch pods for service account info
	podBase := customkube.NewResourceClient(kubeClient, kubeCoreCache)
	pods := v1.NewPodClientWithBase(podBase)

	return newClientset(
		restConfig,
		kubeClient,
		newInputClients(pods),
		newDiscoveryClients(mesh),
	), nil
}
