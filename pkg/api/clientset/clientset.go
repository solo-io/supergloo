package clientset

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	customkube "github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	skkube "github.com/solo-io/supergloo/pkg/api/external/kubernetes/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

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
	install, err := v1.NewInstallClient(clientForCrd(v1.InstallCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := install.Register(); err != nil {
		return nil, err
	}

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

	meshGroup, err := v1.NewMeshGroupClient(clientForCrd(v1.MeshGroupCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := meshGroup.Register(); err != nil {
		return nil, err
	}

	upstream, err := gloov1.NewUpstreamClient(clientForCrd(gloov1.UpstreamCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := upstream.Register(); err != nil {
		return nil, err
	}

	routingRule, err := v1.NewRoutingRuleClient(clientForCrd(v1.RoutingRuleCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := routingRule.Register(); err != nil {
		return nil, err
	}

	securityRule, err := v1.NewSecurityRuleClient(clientForCrd(v1.SecurityRuleCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := securityRule.Register(); err != nil {
		return nil, err
	}

	// ilackarms: should we use Kube secret here? these secrets follow a different format (specific to istio)
	tlsSecret, err := v1.NewTlsSecretClient(&factory.KubeSecretClientFactory{
		Clientset:    kubeClient,
		PlainSecrets: true,
		Cache:        kubeCoreCache,
	})
	if err != nil {
		return nil, err
	}
	if err := tlsSecret.Register(); err != nil {
		return nil, err
	}

	// special resource client wired up to kubernetes pods
	// used by the istio policy syncer to watch pods for service account info
	podBase := customkube.NewResourceClient(kubeClient, kubeCoreCache)
	pods := skkube.NewPodClientWithBase(podBase)

	return newClientset(
		kubeClient,
		newInputClients(install, mesh, meshIngress, meshGroup, upstream, routingRule, securityRule, tlsSecret),
		newDiscoveryClients(pods),
	), nil
}

func IstioFromContext(ctx context.Context) (*IstioClients, error) {
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	crdCache := kube.NewKubeCache(ctx)
	/*
		istio clients
	*/

	rbacConfig, err := rbacv1alpha1.NewRbacConfigClient(&factory.KubeResourceClientFactory{
		Crd:             rbacv1alpha1.RbacConfigCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := rbacConfig.Register(); err != nil {
		return nil, err
	}

	serviceRole, err := rbacv1alpha1.NewServiceRoleClient(&factory.KubeResourceClientFactory{
		Crd:             rbacv1alpha1.ServiceRoleCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := serviceRole.Register(); err != nil {
		return nil, err
	}

	serviceRoleBinding, err := rbacv1alpha1.NewServiceRoleBindingClient(&factory.KubeResourceClientFactory{
		Crd:             rbacv1alpha1.ServiceRoleBindingCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := serviceRoleBinding.Register(); err != nil {
		return nil, err
	}

	meshPolicy, err := policyv1alpha1.NewMeshPolicyClient(&factory.KubeResourceClientFactory{
		Crd:             policyv1alpha1.MeshPolicyCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := meshPolicy.Register(); err != nil {
		return nil, err
	}

	destinationRule, err := v1alpha3.NewDestinationRuleClient(&factory.KubeResourceClientFactory{
		Crd:             v1alpha3.DestinationRuleCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := destinationRule.Register(); err != nil {
		return nil, err
	}

	virtualService, err := v1alpha3.NewVirtualServiceClient(&factory.KubeResourceClientFactory{
		Crd:             v1alpha3.VirtualServiceCrd,
		Cfg:             restConfig,
		SharedCache:     crdCache,
		SkipCrdCreation: true,
	})
	if err != nil {
		return nil, err
	}
	if err := virtualService.Register(); err != nil {
		return nil, err
	}
	return newIstioClients(rbacConfig, serviceRole, serviceRoleBinding, meshPolicy, destinationRule, virtualService), nil
}

type Clientset struct {
	Kube kubernetes.Interface

	// config for supergloo
	Input *inputClients

	// discovery resources from kubernetes
	Discovery *discoveryClients
}

func newClientset(kube kubernetes.Interface, input *inputClients, discovery *discoveryClients) *Clientset {
	return &Clientset{Kube: kube, Input: input, Discovery: discovery}
}

func clientForCrd(crd crd.Crd, restConfig *rest.Config, kubeCache kube.SharedCache) factory.ResourceClientFactory {
	return &factory.KubeResourceClientFactory{Crd: crd, Cfg: restConfig, SharedCache: kubeCache}
}

type inputClients struct {
	Install      v1.InstallClient
	Mesh         v1.MeshClient
	MeshIngress  v1.MeshIngressClient
	MeshGroup    v1.MeshGroupClient
	Upstream     gloov1.UpstreamClient
	RoutingRule  v1.RoutingRuleClient
	SecurityRule v1.SecurityRuleClient
	TlsSecret    v1.TlsSecretClient
}

func newInputClients(install v1.InstallClient, mesh v1.MeshClient, meshIngress v1.MeshIngressClient, meshGroup v1.MeshGroupClient,
	upstream gloov1.UpstreamClient, routingRule v1.RoutingRuleClient, securityRule v1.SecurityRuleClient, tlsSecret v1.TlsSecretClient) *inputClients {
	return &inputClients{Install: install, Mesh: mesh, MeshIngress: meshIngress, MeshGroup: meshGroup, Upstream: upstream,
		RoutingRule: routingRule, SecurityRule: securityRule, TlsSecret: tlsSecret}
}

type discoveryClients struct {
	Pod skkube.PodClient
}

func newDiscoveryClients(pod skkube.PodClient) *discoveryClients {
	return &discoveryClients{Pod: pod}
}

type IstioClients struct {
	RbacConfig         rbacv1alpha1.RbacConfigClient
	ServiceRole        rbacv1alpha1.ServiceRoleClient
	ServiceRoleBinding rbacv1alpha1.ServiceRoleBindingClient
	MeshPolicy         policyv1alpha1.MeshPolicyClient
	DestinationRule    v1alpha3.DestinationRuleClient
	VirtualService     v1alpha3.VirtualServiceClient
}

func newIstioClients(rbacConfig rbacv1alpha1.RbacConfigClient, serviceRole rbacv1alpha1.ServiceRoleClient, serviceRoleBinding rbacv1alpha1.ServiceRoleBindingClient, meshPolicy policyv1alpha1.MeshPolicyClient, destinationRule v1alpha3.DestinationRuleClient, virtualService v1alpha3.VirtualServiceClient) *IstioClients {
	return &IstioClients{RbacConfig: rbacConfig, ServiceRole: serviceRole, ServiceRoleBinding: serviceRoleBinding, MeshPolicy: meshPolicy, DestinationRule: destinationRule, VirtualService: virtualService}
}
