package clientset

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Clientset struct {
	RestConfig *rest.Config

	Kube          kubernetes.Interface
	ApiExtensions clientset.Interface

	Input *inputClients

	Discovery *discoveryClients
}

func newClientset(restConfig *rest.Config, kube kubernetes.Interface, apiExtensions clientset.Interface, input *inputClients, discovery *discoveryClients) *Clientset {
	return &Clientset{RestConfig: restConfig, Kube: kube, ApiExtensions: apiExtensions, Input: input, Discovery: discovery}
}

type discoveryClients struct {
	Mesh v1.MeshClient
}

func newDiscoveryClients(mesh v1.MeshClient) *discoveryClients {
	return &discoveryClients{Mesh: mesh}
}

type inputClients struct {
	Pod        skkube.PodClient
	Deployment skkube.DeploymentClient
	Upstream   gloov1.UpstreamClient
	Secret     gloov1.SecretClient
	//TlsSecret  v1.TlsSecretClient
}

func newInputClients(pod skkube.PodClient, deployment skkube.DeploymentClient, upstream gloov1.UpstreamClient, secret gloov1.SecretClient) *inputClients {
	return &inputClients{Pod: pod, Deployment: deployment, Upstream: upstream, Secret: secret}
}

func clientForCrd(crd crd.Crd, restConfig *rest.Config, kubeCache kube.SharedCache) factory.ResourceClientFactory {
	return &factory.KubeResourceClientFactory{Crd: crd, Cfg: restConfig, SharedCache: kubeCache}
}

//
//// initialize all resource clients here that will share a cache
//func ClientsetFromContext(ctx context.Context) (*Clientset, error) {
//	restConfig, err := kubeutils.GetConfig("", "")
//	if err != nil {
//		return nil, err
//	}
//	kubeClient, err := kubernetes.NewForConfig(restConfig)
//	if err != nil {
//		return nil, err
//	}
//	apiExtsClient, err := clientset.NewForConfig(restConfig)
//	if err != nil {
//		return nil, err
//	}
//	crdCache := kube.NewKubeCache(ctx)
//	kubeCoreCache, err := cache.NewKubeCoreCache(ctx, kubeClient)
//	if err != nil {
//		return nil, err
//	}
//	deploymentCache, err := cache.NewKubeDeploymentCache(ctx, kubeClient)
//	if err != nil {
//		return nil, err
//	}
//	/*
//		supergloo config clients
//	*/
//
//	mesh, err := v1.NewMeshClient(clientForCrd(v1.MeshCrd, restConfig, crdCache))
//	if err != nil {
//		return nil, err
//	}
//	if err := mesh.Register(); err != nil {
//		return nil, err
//	}
//
//	// special resource client wired up to kubernetes pods
//	// used by the istio policy syncer to watch pods for service account info
//	pods := skpod.NewPodClient(kubeClient, kubeCoreCache)
//	deployments := skdeployment.NewDeploymentClient(kubeClient, deploymentCache)
//
//	return newClientset(
//		restConfig,
//		kubeClient,
//		apiExtsClient,
//		newInputClients(pods, deployments, upstream, secret, tlsSecret),
//		newDiscoveryClients(mesh),
//	), nil
//}

//type MeshPolicyClientLoader func() (v1alpha1.MeshPolicyClient, error)

//type IstioClientset struct {
//	MeshPolicies MeshPolicyClientLoader
//}
//
//func newIstioClientset(meshpolicies MeshPolicyClientLoader) *IstioClientset {
//	return &IstioClientset{MeshPolicies: meshpolicies}
//}
//
//func IstioClientsetFromContext(ctx context.Context) (*IstioClientset, error) {
//	restConfig, err := kubeutils.GetConfig("", "")
//	if err != nil {
//		return nil, err
//	}
//	crdCache := kube.NewKubeCache(ctx)
//
//	// the cache should only be registered once
//	registerOnce := &sync.Once{}
//
//	meshPolicyClientLoader := MeshPolicyClientLoader(func() (v1alpha1.MeshPolicyClient, error) {
//		meshPolicyConfig, err := v1alpha1.NewMeshPolicyClient(&factory.KubeResourceClientFactory{
//			Crd:             v1alpha1.MeshPolicyCrd,
//			Cfg:             restConfig,
//			SharedCache:     crdCache,
//			SkipCrdCreation: true,
//		})
//		if err != nil {
//			return nil, err
//		}
//		var registrationErr error
//		registerOnce.Do(func() {
//			registrationErr = meshPolicyConfig.Register()
//		})
//		if registrationErr != nil {
//			return nil, registrationErr
//		}
//
//		return meshPolicyConfig, nil
//	})
//	return newIstioClientset(meshPolicyClientLoader), nil
//
//}
