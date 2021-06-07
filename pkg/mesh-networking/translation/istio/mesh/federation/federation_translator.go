package federation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/config/kube"
	"istio.io/istio/pkg/config/protocol"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

//go:generate mockgen -source ./federation_translator.go -destination mocks/federation_translator.go

const (
	// NOTE(ilackarms): we may want to support federating over non-tls port at some point.
	defaultGatewayProtocol = "TLS"
	DefaultGatewayPortName = "tls"
)

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Mesh.
	// returns nil if no VirtualService or DestinationRule is required for the Mesh (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the istio.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.LocalSnapshot,
		mesh *discoveryv1.Mesh,
		virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
		outputs istio.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx                       context.Context
	destinations              discoveryv1sets.DestinationSet
	virtualServiceTranslator  virtualservice.Translator
	destinationRuleTranslator destinationrule.Translator
}

func NewTranslator(
	ctx context.Context,
	destinations discoveryv1sets.DestinationSet,
	virtualServiceTranslator virtualservice.Translator,
	destinationRuleTranslator destinationrule.Translator,
) Translator {
	return &translator{
		ctx:                       ctx,
		destinations:              destinations,
		virtualServiceTranslator:  virtualServiceTranslator,
		destinationRuleTranslator: destinationRuleTranslator,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	in input.LocalSnapshot,
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	outputs istio.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.Type)
		return
	}
	if virtualMesh == nil || len(virtualMesh.Spec.Meshes) < 2 {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring istio mesh %v which is not federated with other meshes in a VirtualMesh", sets.Key(mesh))
		return
	}
	if len(istioMesh.IngressGateways) < 1 {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring istio mesh %v has no ingress gateway", sets.Key(mesh))
		return
	}
	// TODO(ilackarms): consider supporting multiple ingress gateways or selecting a specific gateway.
	// Currently, we just default to using the first one in the list.
	ingressGateway := istioMesh.IngressGateways[0]

	destinations := DestinationsForMesh(mesh, in.Destinations())

	if len(destinations) == 0 {
		contextutils.LoggerFrom(t.ctx).Debugf("no services found in istio mesh %v", sets.Key(mesh))
		return
	}

	istioCluster := istioMesh.Installation.Cluster
	istioNamespace := istioMesh.Installation.Namespace

	federatedHostnameSuffix := hostutils.GetFederatedHostnameSuffix(virtualMesh.Spec)

	for _, destination := range destinations {
		switch destination.Spec.Type.(type) {
		case *discoveryv1.DestinationSpec_KubeService_:
			t.translateKubeServiceDestination(destination, reporter, virtualMesh, in, outputs, mesh)
		case *discoveryv1.DestinationSpec_ExternalService_:
			// External (non-k8s) service scenario, handled by enterprise networking
			// TODO: Might want to warn user if they have external destinations
			// configured with OSS Gloo Mesh
			continue
		default:
			// Should never happen
			contextutils.LoggerFrom(t.ctx).Debugf("skipping destination %v (only kubeService or externalService supported)")
			continue
		}
	}

	// istio gateway names must be DNS-1123 labels
	// hyphens are legal, dots are not, so we convert here
	gwName := BuildGatewayName(virtualMesh)
	// currently, in non flat-network mode, only TLS-secured cross cluster traffic is supported
	gw := &networkingv1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        gwName,
			Namespace:   istioNamespace,
			ClusterName: istioCluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.Gateway{
			Servers: []*networkingv1alpha3spec.Server{{
				Port: &networkingv1alpha3spec.Port{
					Number:   ingressGateway.TlsContainerPort,
					Protocol: defaultGatewayProtocol,
					Name:     DefaultGatewayPortName,
				},
				Hosts: []string{"*." + federatedHostnameSuffix},
				Tls: &networkingv1alpha3spec.ServerTLSSettings{
					Mode: networkingv1alpha3spec.ServerTLSSettings_AUTO_PASSTHROUGH,
				},
			}},
			Selector: ingressGateway.WorkloadLabels,
		},
	}

	// Append the virtual mesh as a parent to each output resource
	metautils.AppendParent(t.ctx, gw, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())

	outputs.AddGateways(gw)
}

func BuildGatewayName(virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh) string {
	return kubeutils.SanitizeNameV2(
		fmt.Sprintf("%s-%s", virtualMesh.GetRef().GetName(), virtualMesh.GetRef().GetNamespace()),
	)
}

// translateKubeServiceDestination takes a KubeService Destination and adds a ServiceEntry
// to the output for every other cluster in the mesh (not the cluster hosting the KubeService),
// so that the KubeService can be accessed from any cluster in the mesh via a federated hostname.
func (t *translator) translateKubeServiceDestination(
	destination *discoveryv1.Destination,
	reporter reporting.Reporter,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	in input.LocalSnapshot,
	outputs istio.Builder,
	mesh *discoveryv1.Mesh,
) {
	// KubeService scenario
	meshKubeService := destination.Spec.GetKubeService()
	istioMesh := mesh.Spec.GetIstio()

	// Guaranteed to have at least one gateway passed by caller
	ingressGateway := istioMesh.IngressGateways[0]
	federatedHostnameSuffix := hostutils.GetFederatedHostnameSuffix(virtualMesh.Spec)
	serviceEntryIP, err := destinationutils.ConstructUniqueIpForKubeService(meshKubeService.Ref)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(t.ctx).Errorf("unexpected error: failed to generate service entry ip: %v", err)
		return
	}

	federatedHostname := hostutils.BuildFederatedFQDN(
		meshKubeService.GetRef(),
		virtualMesh.Spec,
	)

	remoteIngressTlsPort := make(map[string]uint32)
	var ports []*networkingv1alpha3spec.Port
	for _, port := range destination.Spec.GetKubeService().GetPorts() {
		portName := port.Name
		// fall back to protocol for port name if k8s port name is unpopulated
		if portName == "" {
			portName = port.Protocol
		}
		ports = append(ports, &networkingv1alpha3spec.Port{
			Number:   port.Port,
			Protocol: ConvertKubePortProtocol(port),
			Name:     portName,
		})
		remoteIngressTlsPort[portName] = ingressGateway.ExternalTlsPort
	}

	var workloadEntries []*networkingv1alpha3spec.WorkloadEntry
	// construct a WorkloadEntry for each endpoint (i.e. backing Workload) for the Destination
	for _, endpointSubset := range meshKubeService.EndpointSubsets {
		for _, endpoint := range endpointSubset.Endpoints {
			workloadEntry := &networkingv1alpha3spec.WorkloadEntry{
				Address: ingressGateway.ExternalAddress,
				Ports:   remoteIngressTlsPort,
				Labels:  endpoint.Labels,
			}
			workloadEntries = append(workloadEntries, workloadEntry)
		}
	}

	var remoteDestinationRule *networkingv1alpha3.DestinationRule
	// translate remote resources for all meshes in the virtual mesh
	for _, ref := range virtualMesh.Spec.Meshes {
		groupedMesh, err := in.Meshes().Find(ref)
		if err != nil {
			reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, err)
			continue
		}

		istioMesh := groupedMesh.Spec.GetIstio()
		if istioMesh == nil {
			reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, eris.Errorf("non-istio mesh %v cannot be used in virtual mesh", sets.Key(groupedMesh)))
			continue
		}

		if federatedHostnameSuffix != hostutils.DefaultHostnameSuffix && !istioMesh.SmartDnsProxyingEnabled {
			reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, eris.Errorf(
				"mesh %v does not have smart DNS proxying enabled (hostname suffix can only be specified if all grouped istio meshes have it enabled)",
				sets.Key(groupedMesh),
			))
			continue
		}

		// only translate output resources for client meshes
		if ezkube.RefsMatch(ref, mesh) {
			continue
		}

		se := &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        federatedHostname,
				Namespace:   istioMesh.Installation.Namespace,
				ClusterName: istioMesh.Installation.Cluster,
				Labels:      metautils.TranslatedObjectLabels(),
			},
			Spec: networkingv1alpha3spec.ServiceEntry{
				Addresses:  []string{serviceEntryIP.String()},
				Hosts:      []string{federatedHostname},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
				Endpoints:  workloadEntries,
				Ports:      ports,
			},
		}

		// Append the virtual mesh as a parent to the output service entry
		metautils.AppendParent(t.ctx, se, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())
		outputs.AddServiceEntries(se)

		// Translate VirtualServices for federated Destinations, can be nil
		vs := t.virtualServiceTranslator.Translate(t.ctx, in, destination, istioMesh.Installation, reporter)
		// Append the virtual mesh as a parent to the output virtual service
		metautils.AppendParent(t.ctx, vs, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())
		outputs.AddVirtualServices(vs)

		// Translate DestinationRules for federated Destinations, can be nil
		dr := t.destinationRuleTranslator.Translate(t.ctx, in, destination, istioMesh.Installation, reporter)
		// Append the virtual mesh as a parent to the output destination rule
		metautils.AppendParent(t.ctx, dr, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())
		outputs.AddDestinationRules(dr)
		if remoteDestinationRule == nil {
			remoteDestinationRule = dr
		}

		// Update AppliedFederation data on a Destination's status
		updateDestinationFederationStatus(
			destination,
			federatedHostname,
			ezkube.MakeObjectRef(groupedMesh),
			virtualMesh.Spec.Meshes,
			virtualMesh.Spec.GetFederation().GetFlatNetwork(),
		)
	}

	// translate local resources
	if len(destination.Status.AppliedFederation.GetFederatedToMeshes()) != 0 {
		localServiceEntry, localDestinationRule := translateLocalDestinationFederationResources(destination, mesh, ports, remoteDestinationRule)
		metautils.AppendParent(t.ctx, localServiceEntry, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())
		metautils.AppendParent(t.ctx, localDestinationRule, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())
		outputs.AddServiceEntries(localServiceEntry)
		outputs.AddDestinationRules(localDestinationRule)
	}

}

func translateLocalDestinationFederationResources(
	destination *discoveryv1.Destination,
	destinationMesh *discoveryv1.Mesh,
	serviceEntryPorts []*networkingv1alpha3spec.Port,
	remoteDestinationRule *networkingv1alpha3.DestinationRule,
) (*networkingv1alpha3.ServiceEntry, *networkingv1alpha3.DestinationRule) {
	federatedHostname := destination.Status.AppliedFederation.GetFederatedHostname()
	destinationIstioMesh := destinationMesh.Spec.GetIstio()

	var workloadEntries []*networkingv1alpha3spec.WorkloadEntry
	// construct a WorkloadEntry for each endpoint (i.e. backing Workload) for the Destination
	for _, endpointSubset := range destination.Spec.GetKubeService().EndpointSubsets {
		for _, endpoint := range endpointSubset.Endpoints {

			ports := map[string]uint32{}
			for _, port := range endpointSubset.Ports {
				portName := port.Name
				// fall back to protocol for port name if k8s port name is unpopulated
				if portName == "" {
					portName = port.Protocol
				}
				ports[portName] = port.Port
			}

			workloadEntry := &networkingv1alpha3spec.WorkloadEntry{
				Address: endpoint.IpAddress,
				Ports:   ports,
				Labels:  endpoint.Labels,
			}
			workloadEntries = append(workloadEntries, workloadEntry)
		}
	}

	se := &networkingv1alpha3.ServiceEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:        federatedHostname,
			Namespace:   destinationIstioMesh.Installation.Namespace,
			ClusterName: destinationIstioMesh.Installation.Cluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.ServiceEntry{
			// match the federate hostname
			Hosts: []string{federatedHostname},
			// only export to Gateway workload namespace
			ExportTo:   []string{"."},
			Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
			Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
			Endpoints:  workloadEntries,
			Ports:      serviceEntryPorts,
		},
	}

	var dr *networkingv1alpha3.DestinationRule
	// if the remote DestinationRule is nil, that means no DestinationRule config is required for this federated Destination
	if remoteDestinationRule != nil {
		dr = &networkingv1alpha3.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:        federatedHostname,
				Namespace:   destinationIstioMesh.Installation.Namespace,
				ClusterName: destinationIstioMesh.Installation.Cluster,
				Labels:      metautils.TranslatedObjectLabels(),
			},
			Spec: networkingv1alpha3spec.DestinationRule{
				Host:          federatedHostname,
				Subsets:       remoteDestinationRule.Spec.Subsets,
				TrafficPolicy: remoteDestinationRule.Spec.TrafficPolicy,
			},
		}
	}

	return se, dr
}

func updateDestinationFederationStatus(
	destination *discoveryv1.Destination,
	federatedHostname string,
	mesh *skv2corev1.ObjectRef,
	groupedMeshes []*skv2corev1.ObjectRef,
	flatNetwork bool,
) {
	var federatedToMeshes []*skv2corev1.ObjectRef

	// don't include the mesh of the Destination itself in the list of federated meshes
	for _, ref := range groupedMeshes {
		if ezkube.RefsMatch(ref, mesh) {
			continue
		}
		federatedToMeshes = append(federatedToMeshes, mesh)
	}

	destination.Status.AppliedFederation = &discoveryv1.DestinationStatus_AppliedFederation{
		FederatedHostname: federatedHostname,
		FederatedToMeshes: federatedToMeshes,
		FlatNetwork:       flatNetwork,
	}
}

// DestinationsForMesh returns all Destinations which belong to a given mesh
// exported for use in enterprise
func DestinationsForMesh(
	mesh *discoveryv1.Mesh,
	destinations discoveryv1sets.DestinationSet,
) []*discoveryv1.Destination {
	return destinations.List(func(service *discoveryv1.Destination) bool {
		// Always return external services, they apply to all meshes
		// TODO: Revisit once we have an API for expressing which meshes the
		// external service should be exported to. For now, export to all.
		if service.Spec.Mesh == nil {
			if _, ok := service.Spec.Type.(*discoveryv1.DestinationSpec_ExternalService_); ok {
				// Is External service
				return false
			}
		}
		return !ezkube.RefsMatch(service.Spec.Mesh, mesh)

	})
}

// ConvertKubePortProtocol converts protocol of k8s Service port to application level protocol
// exported for use in enterprise
func ConvertKubePortProtocol(port *discoveryv1.DestinationSpec_KubeService_KubeServicePort) string {
	var appProtocol *string
	if port.AppProtocol != "" {
		appProtocol = pointer.StringPtr(port.AppProtocol)
	}
	convertedProtocol := kube.ConvertProtocol(int32(port.Port), port.Name, corev1.Protocol(port.Protocol), appProtocol)
	if convertedProtocol == protocol.Unsupported {
		return port.Protocol
	}
	return string(convertedProtocol)
}
