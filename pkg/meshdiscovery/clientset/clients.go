package clientset

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/kubeutils"
	sknamespace "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	skpod "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
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
	Pod         skkube.PodClient
	Namespace   skkube.KubeNamespaceClient
	MeshIngress v1.MeshIngressClient
	Upstream    gloov1.UpstreamClient
}

func newInputClients(pod skkube.PodClient, namespace skkube.KubeNamespaceClient, meshIngress v1.MeshIngressClient, upstream gloov1.UpstreamClient) *inputClients {
	return &inputClients{Pod: pod, Namespace: namespace, MeshIngress: meshIngress, Upstream: upstream}
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

	meshIngress, err := v1.NewMeshIngressClient(clientForCrd(v1.MeshIngressCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := meshIngress.Register(); err != nil {
		return nil, err
	}

	/*
		gloo config clients
	*/

	upstream, err := gloov1.NewUpstreamClient(clientForCrd(gloov1.UpstreamCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := upstream.Register(); err != nil {
		return nil, err
	}

	// special resource client wired up to kubernetes pods
	// used by the istio policy syncer to watch pods for service account info
	pods := skpod.NewPodClient(kubeClient, kubeCoreCache)

	namespace := sknamespace.NewNamespaceClient(kubeClient, kubeCoreCache)

	return newClientset(
		restConfig,
		kubeClient,
		newInputClients(pods, namespace, meshIngress, upstream),
		newDiscoveryClients(mesh),
	), nil
}

type IstioClientset struct {
	MeshPolicies v1alpha1.MeshPolicyClient
}

func newIstioClientset(meshpolicies v1alpha1.MeshPolicyClient) *IstioClientset {
	return &IstioClientset{MeshPolicies: meshpolicies}
}

func IstioClientsetFromContext(ctx context.Context) (*IstioClientset, error) {
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	crdCache := kube.NewKubeCache(ctx)
	/*
		istio clients
	*/

	meshPolicyConfig, err := v1alpha1.NewMeshPolicyClient(&factory.KubeResourceClientFactory{
		Crd:             v1alpha1.MeshPolicyCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := meshPolicyConfig.Register(); err != nil {
		return nil, err
	}
	return newIstioClientset(meshPolicyConfig), nil

}
