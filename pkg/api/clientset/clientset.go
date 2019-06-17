package clientset

import (
	"context"
	"fmt"
	"sync"

	"github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/service"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	kubernetes2 "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	accessclient "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/gen/client/access/clientset/versioned"
	specsclient "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/gen/client/specs/clientset/versioned"
	splitclient "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/linkerd"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/smi"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	linkerdv1 "github.com/solo-io/supergloo/pkg/api/external/linkerd/v1"
	promv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	accessv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/smi/access/v1alpha1"
	specsv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/smi/specs/v1alpha1"
	splitv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CrdNotRegisteredErr struct {
	CrdName       string
	OriginalError error
}

func NewCrdNotRegisteredErr(crdName string, originalError error) *CrdNotRegisteredErr {
	return &CrdNotRegisteredErr{CrdName: crdName, OriginalError: originalError}
}

func (e *CrdNotRegisteredErr) Error() string {
	return fmt.Sprintf("cannot create client for %v: crd not registered (err: %v)", e.CrdName, e.OriginalError)
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

	promClient, err := promv1.NewPrometheusConfigClient(prometheus.ResourceClientFactory(kubeClient, kubeCoreCache))
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

	secret, err := gloov1.NewSecretClient(&factory.KubeSecretClientFactory{
		Clientset: kubeClient,
		Cache:     kubeCoreCache,
	})
	if err != nil {
		return nil, err
	}
	if err := secret.Register(); err != nil {
		return nil, err
	}

	settings, err := gloov1.NewSettingsClient(clientForCrd(gloov1.SettingsCrd, restConfig, crdCache))
	if err != nil {
		return nil, err
	}
	if err := settings.Register(); err != nil {
		return nil, err
	}

	// special resource client wired up to kubernetes pods
	// used by the istio policy syncer to watch pods for service account info
	pods := pod.NewPodClient(kubeClient, kubeCoreCache)
	services := service.NewServiceClient(kubeClient, kubeCoreCache)

	return newClientset(
		restConfig,
		kubeClient,
		promClient,
		newSuperglooClients(install, mesh, meshGroup, meshIngress, upstream,
			routingRule, securityRule, tlsSecret, secret, settings),
		newDiscoveryClients(pods, services),
	), nil
}

func IstioFromContext(ctx context.Context) (*IstioClients, error) {
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	crdCache := kube.NewKubeCache(ctx)

	// each resource should only be registered with the cache once
	registerRbacConfigOnce := &sync.Once{}

	rbacConfigClientLoader := func() (rbacv1alpha1.RbacConfigClient, error) {
		rbacConfig, err := rbacv1alpha1.NewRbacConfigClient(clientForCrd(rbacv1alpha1.RbacConfigCrd, restConfig, crdCache))
		if err != nil {
			return nil, err
		}
		registerRbacConfigOnce.Do(func() {
			err = rbacConfig.Register()
		})
		if err != nil {
			return nil, err
		}
		return rbacConfig, nil
	}

	registerServiceRoleOnce := &sync.Once{}
	serviceRoleClientLoader := func() (rbacv1alpha1.ServiceRoleClient, error) {
		serviceRole, err := rbacv1alpha1.NewServiceRoleClient(clientForCrd(rbacv1alpha1.ServiceRoleCrd, restConfig, crdCache))
		if err != nil {
			return nil, err
		}
		registerServiceRoleOnce.Do(func() {
			err = serviceRole.Register()

		})
		if err != nil {
			return nil, err
		}
		return serviceRole, nil
	}

	registerServiceRoleBindingOnce := &sync.Once{}
	serviceRoleBindingClientLoader := func() (rbacv1alpha1.ServiceRoleBindingClient, error) {
		serviceRoleBinding, err := rbacv1alpha1.NewServiceRoleBindingClient(clientForCrd(rbacv1alpha1.ServiceRoleBindingCrd, restConfig, crdCache))
		if err != nil {
			return nil, err
		}
		registerServiceRoleBindingOnce.Do(func() {
			err = serviceRoleBinding.Register()

		})
		if err != nil {
			return nil, err
		}
		return serviceRoleBinding, nil
	}

	registerMeshPolicyOnce := &sync.Once{}
	meshPolicyClientLoader := func() (policyv1alpha1.MeshPolicyClient, error) {
		meshPolicy, err := policyv1alpha1.NewMeshPolicyClient(clientForCrd(policyv1alpha1.MeshPolicyCrd, restConfig, crdCache))
		if err != nil {
			return nil, err
		}
		registerMeshPolicyOnce.Do(func() {
			err = meshPolicy.Register()

		})
		if err != nil {
			return nil, err
		}
		return meshPolicy, nil
	}

	registerDestinationRuleOnce := &sync.Once{}
	destinationRuleClientLoader := func() (v1alpha3.DestinationRuleClient, error) {
		destinationRule, err := v1alpha3.NewDestinationRuleClient(clientForCrd(v1alpha3.DestinationRuleCrd, restConfig, crdCache))
		if err != nil {
			return nil, err
		}
		var registrationErr error
		registerDestinationRuleOnce.Do(func() {
			registrationErr = destinationRule.Register()

		})
		if registrationErr != nil {
			return nil, registrationErr
		}
		return destinationRule, nil
	}

	registerVirtualServiceOnce := &sync.Once{}
	virtualServiceClientLoader := func() (v1alpha3.VirtualServiceClient, error) {
		virtualService, err := v1alpha3.NewVirtualServiceClient(clientForCrd(v1alpha3.VirtualServiceCrd, restConfig, crdCache))
		if err != nil {
			return nil, err
		}
		var registrationErr error
		registerVirtualServiceOnce.Do(func() {
			registrationErr = virtualService.Register()

		})
		if registrationErr != nil {
			return nil, registrationErr
		}
		return virtualService, nil
	}

	return newIstioClients(
		rbacConfigClientLoader,
		serviceRoleClientLoader,
		serviceRoleBindingClientLoader,
		meshPolicyClientLoader,
		destinationRuleClientLoader,
		virtualServiceClientLoader,
	), nil
}

const serviceProfileCrdName = "serviceprofiles.linkerd.io"

func LinkerdFromContext(ctx context.Context) (*LinkerdClients, error) {
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}

	apiExts, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	linkerdClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	var linkerdCache linkerd.Cache
	startCacheOnce := &sync.Once{}

	serviceProfileClientLoader := func() (linkerdv1.ServiceProfileClient, error) {
		if linkerdCache == nil {
			// check that crd is registered before starting cache
			if err := kubeutils.WaitForCrdActive(apiExts, serviceProfileCrdName); err != nil {
				return nil, NewCrdNotRegisteredErr(serviceProfileCrdName, err)
			}

			var cacheInitErr error
			startCacheOnce.Do(func() {
				linkerdCache, cacheInitErr = linkerd.NewLinkerdCache(ctx, linkerdClient)
			})
			if cacheInitErr != nil {
				return nil, cacheInitErr
			}
		}
		baseServiceProfileClient := linkerd.NewResourceClient(linkerdClient, linkerdCache)

		return linkerdv1.NewServiceProfileClientWithBase(baseServiceProfileClient), nil
	}

	return newLinkerdClients(serviceProfileClientLoader), nil
}

const (
	trafficTargetCrdName  = "traffictargets.access.smi-spec.io"
	httpRouteGroupCrdName = "httproutegroups.specs.smi-spec.io"
	trafficSplitCrdName   = "split.smi-spec.io"
)

func SMIFromContext(ctx context.Context) (*SMIClients, error) {
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	/*
		smi clients
	*/
	accessClient, err := accessclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	specsClient, err := specsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	splitClient, err := splitclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	apiExts, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	var smiCache smi.Cache
	startCacheOnce := &sync.Once{}
	initCache := func() error {
		if smiCache == nil {
			// check that smi crds are registered before starting cache
			if err := kubeutils.WaitForCrdActive(apiExts, trafficTargetCrdName); err != nil {
				return NewCrdNotRegisteredErr(trafficTargetCrdName, err)
			}
			if err := kubeutils.WaitForCrdActive(apiExts, httpRouteGroupCrdName); err != nil {
				return NewCrdNotRegisteredErr(httpRouteGroupCrdName, err)
			}
			if err := kubeutils.WaitForCrdActive(apiExts, trafficSplitCrdName); err != nil {
				return NewCrdNotRegisteredErr(trafficSplitCrdName, err)
			}
			startCacheOnce.Do(func() {
				smiCache, err = smi.NewSMICache(ctx, accessClient, specsClient, splitClient)
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	trafficTargetClientLoader := func() (accessv1alpha1.TrafficTargetClient, error) {
		if err := initCache(); err != nil {
			return nil, err
		}

		return smi.NewTrafficTargetClient(accessClient, smiCache), nil

	}
	httpRouteGroupClientLoader := func() (specsv1alpha1.HTTPRouteGroupClient, error) {
		if err := initCache(); err != nil {
			return nil, err
		}

		return smi.NewHTTPRouteGroupClient(specsClient, smiCache), nil
	}

	trafficSplitClientLoader := func() (splitv1alpha1.TrafficSplitClient, error) {
		if err := initCache(); err != nil {
			return nil, err
		}

		return smi.NewTrafficSplitClient(splitClient, smiCache), nil
	}

	return newSMIClients(trafficTargetClientLoader, httpRouteGroupClientLoader, trafficSplitClientLoader), nil
}

type Clientset struct {
	RestConfig *rest.Config

	Kube kubernetes.Interface

	Prometheus promv1.PrometheusConfigClient

	// config for supergloo
	Supergloo *SuperglooClients

	// discovery resources from kubernetes
	Discovery *discoveryClients
}

func newClientset(restConfig *rest.Config, kube kubernetes.Interface, prometheus promv1.PrometheusConfigClient, input *SuperglooClients, discovery *discoveryClients) *Clientset {
	return &Clientset{RestConfig: restConfig, Kube: kube, Prometheus: prometheus, Supergloo: input, Discovery: discovery}
}

func clientForCrd(crd crd.Crd, restConfig *rest.Config, kubeCache kube.SharedCache) factory.ResourceClientFactory {
	return &factory.KubeResourceClientFactory{Crd: crd, Cfg: restConfig, SharedCache: kubeCache, SkipCrdCreation: true}
}

type SuperglooClients struct {
	Install      v1.InstallClient
	Mesh         v1.MeshClient
	MeshGroup    v1.MeshGroupClient
	MeshIngress  v1.MeshIngressClient
	Upstream     gloov1.UpstreamClient
	RoutingRule  v1.RoutingRuleClient
	SecurityRule v1.SecurityRuleClient
	TlsSecret    v1.TlsSecretClient
	Secret       gloov1.SecretClient
	Settings     gloov1.SettingsClient
}

func newSuperglooClients(install v1.InstallClient, mesh v1.MeshClient, meshGroup v1.MeshGroupClient,
	meshIngress v1.MeshIngressClient, upstream gloov1.UpstreamClient, routingRule v1.RoutingRuleClient,
	securityRule v1.SecurityRuleClient, tlsSecret v1.TlsSecretClient, secret gloov1.SecretClient, settings gloov1.SettingsClient) *SuperglooClients {
	return &SuperglooClients{Install: install, Mesh: mesh, MeshGroup: meshGroup, MeshIngress: meshIngress,
		Upstream: upstream, RoutingRule: routingRule, SecurityRule: securityRule, TlsSecret: tlsSecret, Secret: secret, Settings: settings}
}

type discoveryClients struct {
	Pod     kubernetes2.PodClient
	Service kubernetes2.ServiceClient
}

func newDiscoveryClients(pod kubernetes2.PodClient, service kubernetes2.ServiceClient) *discoveryClients {
	return &discoveryClients{Pod: pod, Service: service}
}

type RbacConfigClientLoader func() (rbacv1alpha1.RbacConfigClient, error)
type ServiceRoleClientLoader func() (rbacv1alpha1.ServiceRoleClient, error)
type ServiceRoleBindingClientLoader func() (rbacv1alpha1.ServiceRoleBindingClient, error)
type MeshPolicyClientLoader func() (policyv1alpha1.MeshPolicyClient, error)
type DestinationRuleClientLoader func() (v1alpha3.DestinationRuleClient, error)
type VirtualServiceClientLoader func() (v1alpha3.VirtualServiceClient, error)

type IstioClients struct {
	RbacConfig         RbacConfigClientLoader
	ServiceRole        ServiceRoleClientLoader
	ServiceRoleBinding ServiceRoleBindingClientLoader
	MeshPolicy         MeshPolicyClientLoader
	DestinationRule    DestinationRuleClientLoader
	VirtualService     VirtualServiceClientLoader
}

func newIstioClients(rbacConfig RbacConfigClientLoader, serviceRole ServiceRoleClientLoader, serviceRoleBinding ServiceRoleBindingClientLoader, meshPolicy MeshPolicyClientLoader, destinationRule DestinationRuleClientLoader, virtualService VirtualServiceClientLoader) *IstioClients {
	return &IstioClients{RbacConfig: rbacConfig, ServiceRole: serviceRole, ServiceRoleBinding: serviceRoleBinding, MeshPolicy: meshPolicy, DestinationRule: destinationRule, VirtualService: virtualService}
}

type ServiceProfileClientLoader func() (linkerdv1.ServiceProfileClient, error)

type LinkerdClients struct {
	ServiceProfile ServiceProfileClientLoader
}

func newLinkerdClients(serviceProfile ServiceProfileClientLoader) *LinkerdClients {
	return &LinkerdClients{ServiceProfile: serviceProfile}
}

type TrafficTargetClientLoader func() (accessv1alpha1.TrafficTargetClient, error)
type HTTPRouteGroupClientLoader func() (specsv1alpha1.HTTPRouteGroupClient, error)
type TrafficSplitClientLoader func() (splitv1alpha1.TrafficSplitClient, error)

type SMIClients struct {
	TrafficTarget  TrafficTargetClientLoader
	HTTPRouteGroup HTTPRouteGroupClientLoader
	TrafficSplit   TrafficSplitClientLoader
}

func newSMIClients(trafficTarget TrafficTargetClientLoader, HTTPRouteGroup HTTPRouteGroupClientLoader, trafficSplit TrafficSplitClientLoader) *SMIClients {
	return &SMIClients{TrafficTarget: trafficTarget, HTTPRouteGroup: HTTPRouteGroup, TrafficSplit: trafficSplit}
}
