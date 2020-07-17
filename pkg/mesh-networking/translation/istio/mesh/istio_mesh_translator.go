package mesh

import (
	"context"
	"fmt"
	"hash/fnv"
	"net"

	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
	istiov1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/common/defaults"
	"github.com/solo-io/smh/pkg/mesh-networking/reporting"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/protoutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/security/proto/envoy/config/filter/network/tcp_cluster_rewrite/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// TODO(ilackarms): allow users to configure these fields, or make them discoverable from the target istio mesh.
	DefaultGatewayPort     = uint32(15443) // https://istio.io/docs/ops/deployment/requirements/#ports-used-by-istio
	DefaultGatewayProtocol = "TLS"
	DefaultGatewayPortName = "tls"
	// "istio": "ingressgateway" is a known string pair to Istio- it's semantically meaningful but unfortunately not exported from anywhere
	// their ingress gateway is hardcoded in their own implementation to have this label
	// https://github.com/istio/istio/blob/4e27ddc64f6a12e622c4cd5c836f5d7edf94e971/istioctl/cmd/describe.go#L1138
	DefaultGatewayWorkloadLabels = map[string]string{
		"istio": "ingressgateway",
	}

	envoySniClusterFilterName        = "envoy.filters.network.sni_cluster"
	envoyTcpClusterRewriteFilterName = "envoy.filters.network.tcp_cluster_rewrite"

	// We must generate IPs to use for Service Entries. for now we simply generate them from this subnet.
	// https://preliminary.istio.io/docs/setup/install/multicluster/gateways/#configure-the-example-services
	//
	// TODO(ilackarms): allow this to be inferred, configured by the user, or remove when
	// istio supports creating service entries without assigning IPs.
	ipAssignableSubnet = "240.0.0.0/4"
)

// outputs of translating a single Mesh
type Outputs struct {
	Gateway          *istiov1alpha3.Gateway
	EnvoyFilter      *istiov1alpha3.EnvoyFilter
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
		mesh *discoveryv1alpha1.Mesh,
		reporter reporting.Reporter,
	) Outputs
}

type translator struct {
	ctx            context.Context
	clusterDomains hostutils.ClusterDomainRegistry
}

func NewTranslator(ctx context.Context, clusterDomains hostutils.ClusterDomainRegistry, pluginFactory plugins.Factory) Translator {
	return &translator{ctx: ctx, clusterDomains: clusterDomains}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	in input.Snapshot,
	mesh *discoveryv1alpha1.Mesh,
	reporter reporting.Reporter,
) Outputs {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return Outputs{}
	}
	if mesh.Status.AppliedVirtualMesh == nil || len(mesh.Status.AppliedVirtualMesh.Spec.Meshes) < 2 {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring istio mesh %v which is not federated with other meshes in a virtual mesh", sets.Key(mesh))
		return Outputs{}
	}

	meshServices := servicesForMesh(mesh, in.MeshServices())

	if meshServices.Length() == 0 {
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
	for _, meshService := range meshServices.List() {
		meshKubeService := meshService.Spec.GetKubeService()
		if meshKubeService == nil {
			// should never happen
			contextutils.LoggerFrom(t.ctx).Debugf("skipping mesh service %v (only kube types supported)", err)
			continue
		}

		serviceEntryIp, err := constructUniqueIp(meshService)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(t.ctx).Errorf("unexpected error: failed to generate service entry ip: %v", err)
			continue
		}
		federatedHostname := t.clusterDomains.GetServiceGlobalFQDN(meshKubeService.GetRef())

		// add the hostname to the set
		federatedHostnames = append(federatedHostnames, federatedHostname)

		endpointPorts := make(map[string]uint32)
		var ports []*istiov1alpha3spec.Port
		for _, port := range meshService.Spec.GetKubeService().GetPorts() {
			ports = append(ports, &istiov1alpha3spec.Port{
				Number:   port.Port,
				Protocol: port.Protocol,
				Name:     port.Name,
			})
			endpointPorts[port.Name] = istioMesh.IngressGateways[0].TlsServicePort
		}

		endpoints := []*istiov1alpha3spec.WorkloadEntry{{
			Address: istioMesh.IngressGateways[0].ExternalAddress,
			Ports:   endpointPorts,
		}}

		// list all client meshes
		for _, ref := range mesh.Status.AppliedVirtualMesh.Spec.Meshes {
			if ezkube.RefsMatch(ref, mesh) {
				continue
			}
			clientMesh, err := in.Meshes().Find(ref)
			if err != nil {
				reporter.ReportVirtualMesh(mesh, mesh.Status.AppliedVirtualMesh.Ref, err)
				continue
			}
			clientIstio := clientMesh.Spec.GetIstio()
			if clientIstio == nil {
				reporter.ReportVirtualMesh(mesh, mesh.Status.AppliedVirtualMesh.Ref, eris.Errorf("non-istio mesh %v cannot be used in virtual mesh", sets.Key(clientMesh)))
				continue
			}

			se := &istiov1alpha3.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:        federatedHostname,
					Namespace:   clientIstio.Installation.Namespace,
					ClusterName: clientIstio.Installation.Cluster,
					Labels:      metautils.TranslatedObjectLabels(),
				},
				Spec: istiov1alpha3spec.ServiceEntry{
					Addresses:  []string{serviceEntryIp.String()},
					Hosts:      []string{federatedHostname},
					Location:   istiov1alpha3spec.ServiceEntry_MESH_INTERNAL,
					Resolution: istiov1alpha3spec.ServiceEntry_DNS,
					Endpoints:  endpoints,
					Ports:      ports,
				},
			}

			dr := &istiov1alpha3.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:        federatedHostname,
					Namespace:   clientIstio.Installation.Namespace,
					ClusterName: clientIstio.Installation.Cluster,
					Labels:      metautils.TranslatedObjectLabels(),
				},
				Spec: istiov1alpha3spec.DestinationRule{
					Host: federatedHostname,
					TrafficPolicy: &istiov1alpha3spec.TrafficPolicy{
						Tls: &istiov1alpha3spec.ClientTLSSettings{
							// TODO this won't work with other mesh types https://github.com/solo-io/service-mesh-hub/issues/242
							Mode: istiov1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}

			serviceEntries.Insert(se)
			destinationRules.Insert(dr)
		}
	}

	gw := &istiov1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   istioNamespace,
			Name:        mesh.Name,
			ClusterName: istioCluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: istiov1alpha3spec.Gateway{
			Servers: []*istiov1alpha3spec.Server{{
				Port: &istiov1alpha3spec.Port{
					Number:   DefaultGatewayPort,
					Protocol: DefaultGatewayProtocol,
					Name:     DefaultGatewayPortName,
				},
				Hosts: federatedHostnames,
				Tls: &istiov1alpha3spec.ServerTLSSettings{
					Mode: istiov1alpha3spec.ServerTLSSettings_AUTO_PASSTHROUGH,
				},
			}},
			Selector: DefaultGatewayWorkloadLabels,
		},
	}

	ef := &istiov1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   istioNamespace,
			Name:        mesh.Name,
			ClusterName: istioCluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: istiov1alpha3spec.EnvoyFilter{
			WorkloadSelector: &istiov1alpha3spec.WorkloadSelector{
				Labels: DefaultGatewayWorkloadLabels,
			},
			ConfigPatches: []*istiov1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch{{
				ApplyTo: istiov1alpha3spec.EnvoyFilter_NETWORK_FILTER,
				Match: &istiov1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch{
					Context: istiov1alpha3spec.EnvoyFilter_GATEWAY,
					ObjectTypes: &istiov1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
						Listener: &istiov1alpha3spec.EnvoyFilter_ListenerMatch{
							PortNumber: DefaultGatewayPort,
							FilterChain: &istiov1alpha3spec.EnvoyFilter_ListenerMatch_FilterChainMatch{
								Filter: &istiov1alpha3spec.EnvoyFilter_ListenerMatch_FilterMatch{
									Name: envoySniClusterFilterName,
								},
							},
						}},
				},
				Patch: &istiov1alpha3spec.EnvoyFilter_Patch{
					Operation: istiov1alpha3spec.EnvoyFilter_Patch_INSERT_AFTER,
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

func constructUniqueIp(meshService *discoveryv1alpha1.MeshService) (net.IP, error) {
	ip, cidr, err := net.ParseCIDR(ipAssignableSubnet)
	if err != nil {
		return nil, err
	}
	if len(ip) != 4 {
		return nil, eris.Errorf("unexpected length for cidr IP: %v", len(ip))
	}

	ip = ip.Mask(cidr.Mask)

	h := fnv.New32()
	if _, err := h.Write([]byte(meshService.Name)); err != nil {
		return nil, err
	}
	if _, err := h.Write([]byte(meshService.Namespace)); err != nil {
		return nil, err
	}
	if _, err := h.Write([]byte(meshService.ClusterName)); err != nil {
		return nil, err
	}
	hash := h.Sum32()
	var hashedIP net.IP = []byte{
		byte(hash),
		byte(hash >> 8),
		byte(hash >> 16),
		byte(hash >> 24),
	}
	hashedIP.Mask(cidr.Mask)

	for i := range hashedIP {
		hashedIP[i] = hashedIP[i] | ip[i]
	}

	return hashedIP, nil
}

func servicesForMesh(
	mesh *discoveryv1alpha1.Mesh,
	allMeshServices v1alpha1sets.MeshServiceSet,
) v1alpha1sets.MeshServiceSet {
	servicesForMesh := v1alpha1sets.NewMeshServiceSet()
	for _, service := range allMeshServices.List() {
		if ezkube.RefsMatch(service.Spec.Mesh, mesh) {
			servicesForMesh.Insert(service)
		}
	}
	return servicesForMesh
}

func buildTcpRewritePatch(clusterName, clusterDomain string) (*types.Struct, error) {
	if clusterDomain == "" {
		clusterDomain = defaults.DefaultClusterDomain
	}
	tcpRewrite, err := protoutils.MessageToGolangStruct(&v2alpha1.TcpClusterRewrite{
		ClusterPattern:     fmt.Sprintf("\\.%s$", clusterName),
		ClusterReplacement: ".svc." + clusterDomain,
	})
	if err != nil {
		return nil, err
	}
	return protoutils.MessageToGogoStruct(&envoy_api_v2_listener.Filter{
		Name: envoyTcpClusterRewriteFilterName,
		ConfigType: &envoy_api_v2_listener.Filter_Config{
			Config: tcpRewrite,
		},
	})
}
