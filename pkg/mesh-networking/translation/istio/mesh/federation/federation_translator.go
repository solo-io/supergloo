package federation

import (
	"context"
	"fmt"

	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
	istiov1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/meshserviceutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/protoutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/security/proto/envoy/config/filter/network/tcp_cluster_rewrite/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// NOTE(ilackarms): we may want to support federating over non-tls port at some point.
	defaultGatewayProtocol = "TLS"
	defaultGatewayPortName = "tls"

	envoySniClusterFilterName        = "envoy.filters.network.sni_cluster"
	envoyTcpClusterRewriteFilterName = "envoy.filters.network.tcp_cluster_rewrite"
)

// outputs of translating a single Mesh
type Outputs struct {
	Gateway          *networkingv1alpha3.Gateway
	EnvoyFilter      *networkingv1alpha3.EnvoyFilter
	DestinationRules istiov1alpha3sets.DestinationRuleSet
	ServiceEntries   istiov1alpha3sets.ServiceEntrySet
}

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Mesh.
	// returns nil if no VirtualService or DestinationRule is required for the Mesh (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
		reporter reporting.Reporter,
	) Outputs
}

type translator struct {
	ctx            context.Context
	clusterDomains hostutils.ClusterDomainRegistry
}

func NewTranslator(ctx context.Context, clusterDomains hostutils.ClusterDomainRegistry) Translator {
	return &translator{ctx: ctx, clusterDomains: clusterDomains}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	in input.Snapshot,
	mesh *discoveryv1alpha2.Mesh,
	virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
	reporter reporting.Reporter,
) Outputs {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return Outputs{}
	}
	if virtualMesh == nil || len(virtualMesh.Spec.Meshes) < 2 {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring istio mesh %v which is not federated with other meshes in a virtual mesh", sets.Key(mesh))
		return Outputs{}
	}
	if len(istioMesh.IngressGateways) < 1 {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring istio mesh %v has no ingress gateway", sets.Key(mesh))
		return Outputs{}
	}
	// TODO(ilackarms): consider supporting multiple ingress gateways or selecting a specific gateway.
	// Currently, we just default to using the first one in the list.
	ingressGateway := istioMesh.IngressGateways[0]

	meshServices := servicesForMesh(mesh, in.MeshServices())

	if len(meshServices) == 0 {
		contextutils.LoggerFrom(t.ctx).Debugf("no services found in istio mesh %v", sets.Key(mesh))
		return Outputs{}
	}

	istioCluster := istioMesh.Installation.Cluster

	kubeCluster, err := in.KubernetesClusters().Find(&v1.ObjectRef{
		Name:      istioCluster,
		Namespace: defaults.GetPodNamespace(),
	})
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("internal error: cluster %v for istio mesh %v not found", istioCluster, sets.Key(mesh))
		return Outputs{}
	}

	istioNamespace := istioMesh.Installation.Namespace

	tcpRewritePatch, err := buildTcpRewritePatch(
		istioCluster,
		kubeCluster.Spec.ClusterDomain,
	)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(t.ctx).DPanicf("failed generating tcp rewrite patch: %v", err)
		return Outputs{}
	}

	destinationRules := istiov1alpha3sets.NewDestinationRuleSet()

	serviceEntries := istiov1alpha3sets.NewServiceEntrySet()

	var federatedHostnames []string
	for _, meshService := range meshServices {
		meshKubeService := meshService.Spec.GetKubeService()
		if meshKubeService == nil {
			// should never happen
			contextutils.LoggerFrom(t.ctx).Debugf("skipping mesh service %v (only kube types supported)", err)
			continue
		}

		serviceEntryIp, err := meshserviceutils.ConstructUniqueIpForKubeService(meshKubeService.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(t.ctx).Errorf("unexpected error: failed to generate service entry ip: %v", err)
			continue
		}
		federatedHostname := t.clusterDomains.GetServiceGlobalFQDN(meshKubeService.GetRef())

		// add the hostname to the set
		federatedHostnames = append(federatedHostnames, federatedHostname)

		endpointPorts := make(map[string]uint32)
		var ports []*networkingv1alpha3spec.Port
		for _, port := range meshService.Spec.GetKubeService().GetPorts() {
			ports = append(ports, &networkingv1alpha3spec.Port{
				Number:   port.Port,
				Protocol: port.Protocol,
				Name:     port.Name,
			})
			endpointPorts[port.Name] = ingressGateway.ExternalTlsPort
		}

		endpoints := []*networkingv1alpha3spec.WorkloadEntry{{
			Address: ingressGateway.ExternalAddress,
			Ports:   endpointPorts,
		}}

		// list all client meshes
		for _, ref := range virtualMesh.Spec.Meshes {
			if ezkube.RefsMatch(ref, mesh) {
				continue
			}
			clientMesh, err := in.Meshes().Find(ref)
			if err != nil {
				reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, err)
				continue
			}
			clientIstio := clientMesh.Spec.GetIstio()
			if clientIstio == nil {
				reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, eris.Errorf("non-istio mesh %v cannot be used in virtual mesh", sets.Key(clientMesh)))
				continue
			}

			se := &networkingv1alpha3.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:        federatedHostname,
					Namespace:   clientIstio.Installation.Namespace,
					ClusterName: clientIstio.Installation.Cluster,
					Labels:      metautils.TranslatedObjectLabels(),
				},
				Spec: networkingv1alpha3spec.ServiceEntry{
					Addresses:  []string{serviceEntryIp.String()},
					Hosts:      []string{federatedHostname},
					Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
					Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
					Endpoints:  endpoints,
					Ports:      ports,
				},
			}

			dr := &networkingv1alpha3.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:        federatedHostname,
					Namespace:   clientIstio.Installation.Namespace,
					ClusterName: clientIstio.Installation.Cluster,
					Labels:      metautils.TranslatedObjectLabels(),
				},
				Spec: networkingv1alpha3spec.DestinationRule{
					Host: federatedHostname,
					TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
						Tls: &networkingv1alpha3spec.ClientTLSSettings{
							// TODO this won't work with other mesh types https://github.com/solo-io/service-mesh-hub/issues/242
							Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}

			serviceEntries.Insert(se)
			destinationRules.Insert(dr)
		}
	}

	gw := &networkingv1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%v.%v", virtualMesh.Ref.Name, virtualMesh.Ref.Namespace),
			Namespace:   istioNamespace,
			ClusterName: istioCluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.Gateway{
			Servers: []*networkingv1alpha3spec.Server{{
				Port: &networkingv1alpha3spec.Port{
					Number:   ingressGateway.TlsContainerPort,
					Protocol: defaultGatewayProtocol,
					Name:     defaultGatewayPortName,
				},
				Hosts: federatedHostnames,
				Tls: &networkingv1alpha3spec.ServerTLSSettings{
					Mode: networkingv1alpha3spec.ServerTLSSettings_AUTO_PASSTHROUGH,
				},
			}},
			Selector: ingressGateway.WorkloadLabels,
		},
	}

	ef := &networkingv1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%v.%v", virtualMesh.Ref.Name, virtualMesh.Ref.Namespace),
			Namespace:   istioNamespace,
			ClusterName: istioCluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.EnvoyFilter{
			WorkloadSelector: &networkingv1alpha3spec.WorkloadSelector{
				Labels: ingressGateway.WorkloadLabels,
			},
			ConfigPatches: []*networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch{{
				ApplyTo: networkingv1alpha3spec.EnvoyFilter_NETWORK_FILTER,
				Match: &networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch{
					Context: networkingv1alpha3spec.EnvoyFilter_GATEWAY,
					ObjectTypes: &networkingv1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
						Listener: &networkingv1alpha3spec.EnvoyFilter_ListenerMatch{
							PortNumber: ingressGateway.TlsContainerPort,
							FilterChain: &networkingv1alpha3spec.EnvoyFilter_ListenerMatch_FilterChainMatch{
								Filter: &networkingv1alpha3spec.EnvoyFilter_ListenerMatch_FilterMatch{
									Name: envoySniClusterFilterName,
								},
							},
						}},
				},
				Patch: &networkingv1alpha3spec.EnvoyFilter_Patch{
					Operation: networkingv1alpha3spec.EnvoyFilter_Patch_INSERT_AFTER,
					Value:     tcpRewritePatch,
				},
			}},
		},
	}

	return Outputs{
		Gateway:          gw,
		EnvoyFilter:      ef,
		DestinationRules: destinationRules,
		ServiceEntries:   serviceEntries,
	}
}

func servicesForMesh(
	mesh *discoveryv1alpha2.Mesh,
	allMeshServices discoveryv1alpha2sets.MeshServiceSet,
) []*discoveryv1alpha2.MeshService {
	return allMeshServices.List(func(service *discoveryv1alpha2.MeshService) bool {
		return !ezkube.RefsMatch(service.Spec.Mesh, mesh)
	})
}

func buildTcpRewritePatch(clusterName, clusterDomain string) (*types.Struct, error) {
	if clusterDomain == "" {
		clusterDomain = defaults.DefaultClusterDomain
	}
	tcpRewrite, err := protoutils.GogoMessageToGolangStruct(&v2alpha1.TcpClusterRewrite{
		ClusterPattern:     fmt.Sprintf("\\.%s$", clusterName),
		ClusterReplacement: ".svc." + clusterDomain,
	})
	if err != nil {
		return nil, err
	}
	return protoutils.GogoMessageToGogoStruct(&envoy_api_v2_listener.Filter{
		Name: envoyTcpClusterRewriteFilterName,
		ConfigType: &envoy_api_v2_listener.Filter_Config{
			Config: tcpRewrite,
		},
	})
}
