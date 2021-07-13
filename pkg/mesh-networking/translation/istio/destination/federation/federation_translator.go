package federation

import (
	"context"
	"net"
	"strings"

	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
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

// the Federation translator translates a Destination into a ServiceEntry, VirtualService, and DestinationRule
type Translator interface {

	// Translate translates a ServiceEntry, VirtualService and DestinationRule for the given Destination using the data in status.AppliedFederation.
	// returns nil if no VirtualService or DestinationRule is required for the Mesh (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.LocalSnapshot,
		destination *discoveryv1.Destination,
		reporter reporting.Reporter,
	) (
		[]*networkingv1alpha3.ServiceEntry,
		[]*networkingv1alpha3.VirtualService,
		[]*networkingv1alpha3.DestinationRule,
	)
}

type translator struct {
	ctx                       context.Context
	virtualServiceTranslator  virtualservice.Translator
	destinationRuleTranslator destinationrule.Translator
}

func NewTranslator(
	ctx context.Context,
	virtualServiceTranslator virtualservice.Translator,
	destinationRuleTranslator destinationrule.Translator,
) Translator {
	return &translator{
		ctx:                       ctx,
		virtualServiceTranslator:  virtualServiceTranslator,
		destinationRuleTranslator: destinationRuleTranslator,
	}
}

// Translate translates a ServiceEntry, VirtualService and DestinationRule for the given Destination using the data in status.AppliedFederation.
func (t *translator) Translate(
	in input.LocalSnapshot,
	destination *discoveryv1.Destination,
	reporter reporting.Reporter,
) (
	[]*networkingv1alpha3.ServiceEntry,
	[]*networkingv1alpha3.VirtualService,
	[]*networkingv1alpha3.DestinationRule,
) {
	// nothing to translate if this Destination is not federated to any external meshes
	if len(destination.Status.AppliedFederation.GetFederatedToMeshes()) == 0 {
		return nil, nil, nil
	}

	// KubeService scenario
	kubeService := destination.Spec.GetKubeService()
	appliedFederation := destination.Status.AppliedFederation
	destinationMesh, err := in.Meshes().Find(destination.Spec.GetMesh())
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("Could not find parent Mesh %v for Destination %v", destination.Spec.GetMesh(), ezkube.MakeObjectRef(destination))
		return nil, nil, nil
	}

	// skip translation if:
	//     1. Destination is not a Kubernetes Service
	//     2. Destination is not federated by a VirtualMesh
	//     3. Destination's parent Mesh is not Istio
	if kubeService == nil || appliedFederation == nil || destinationMesh.Spec.GetIstio() == nil {
		return nil, nil, nil
	}

	destinationVirtualMesh, err := in.VirtualMeshes().Find(destination.Status.AppliedFederation.GetVirtualMeshRef())
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("Could not find parent VirtualMesh %v for Destination %v", destination.Status.AppliedFederation.GetVirtualMeshRef(), ezkube.MakeObjectRef(destination))
		return nil, nil, nil
	}

	// translate ServiceEntry template
	remoteServiceEntryTemplate, err := t.translateRemoteServiceEntryTemplate(destination, destinationMesh)
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("Encountered error while translating ServiceEntry template for Destination %v: %v", ezkube.MakeObjectRef(destination), err)
		return nil, nil, nil
	}

	var serviceEntries []*networkingv1alpha3.ServiceEntry
	var virtualServices []*networkingv1alpha3.VirtualService
	var destinationRules []*networkingv1alpha3.DestinationRule

	var remoteDestinationRule *networkingv1alpha3.DestinationRule
	// translate remote resources
	for _, meshRef := range destination.Status.AppliedFederation.GetFederatedToMeshes() {
		remoteMesh, err := in.Meshes().Find(meshRef)
		if err != nil {
			contextutils.LoggerFrom(t.ctx).Errorf("Could not find Mesh %v that Destination %v is federated to", meshRef, ezkube.MakeObjectRef(destination))
			continue
		}

		serviceEntry, virtualService, destinationRule := t.translateForRemoteMesh(
			destination,
			destinationVirtualMesh,
			in,
			remoteMesh,
			remoteServiceEntryTemplate,
			reporter,
		)

		// Append the VirtualMesh as a parent to the outputs
		metautils.AppendParent(t.ctx, serviceEntry, destination.Status.AppliedFederation.GetVirtualMeshRef(), networkingv1.VirtualMesh{}.GVK())
		metautils.AppendParent(t.ctx, virtualService, destination.Status.AppliedFederation.GetVirtualMeshRef(), networkingv1.VirtualMesh{}.GVK())
		metautils.AppendParent(t.ctx, destinationRule, destination.Status.AppliedFederation.GetVirtualMeshRef(), networkingv1.VirtualMesh{}.GVK())

		serviceEntries = append(serviceEntries, serviceEntry)
		virtualServices = append(virtualServices, virtualService)
		destinationRules = append(destinationRules, destinationRule)

		// take a reference to any translated remote DestinationRule so that we can copy over any necessary fields for the local DestinationRule for the federated FQDN
		// this avoids re-translating the DestinationRule
		if remoteDestinationRule == nil {
			remoteDestinationRule = destinationRule
		}
	}

	// translate local resources
	localServiceEntry, localDestinationRule, err := t.translateForLocalMesh(
		destination,
		destinationMesh,
		remoteServiceEntryTemplate,
		remoteDestinationRule,
	)
	if err != nil {
		reporter.ReportVirtualMeshToDestination(destination, destinationVirtualMesh, err)
		return nil, nil, nil
	}

	// Append the VirtualMesh as a parent to the outputs
	metautils.AppendParent(t.ctx, localServiceEntry, destination.Status.AppliedFederation.GetVirtualMeshRef(), networkingv1.VirtualMesh{}.GVK())
	metautils.AppendParent(t.ctx, localDestinationRule, destination.Status.AppliedFederation.GetVirtualMeshRef(), networkingv1.VirtualMesh{}.GVK())
	serviceEntries = append(serviceEntries, localServiceEntry)
	destinationRules = append(destinationRules, localDestinationRule)

	return serviceEntries, virtualServices, destinationRules
}

// translate the ServiceEntry template that must exist on all meshes this Destination is federated to
func (t *translator) translateRemoteServiceEntryTemplate(
	destination *discoveryv1.Destination,
	destinationMesh *discoveryv1.Mesh,
) (*networkingv1alpha3.ServiceEntry, error) {
	kubeService := destination.Spec.GetKubeService()

	if len(destinationMesh.Status.AppliedEastWestIngressGateways) < 1 {
		return nil, eris.Errorf("istio mesh %v has no applied east west ingress gateways", sets.Key(destinationMesh))
	}

	serviceEntryIP, err := destinationutils.ConstructUniqueIpForKubeService(kubeService.GetRef())
	if err != nil {
		// should never happen
		return nil, eris.Errorf("unexpected error: failed to generate service entry ip: %v", err)
	}

	var sePorts []*networkingv1alpha3spec.Port
	workloadEntryPortMapping := make(map[string]uint32)
	for _, port := range kubeService.GetPorts() {
		portName := port.Name
		// fall back to protocol for port name if k8s port name is unpopulated
		if portName == "" {
			portName = port.Protocol
		}
		sePorts = append(sePorts, &networkingv1alpha3spec.Port{
			Number:   port.Port,
			Protocol: ConvertKubePortProtocol(port),
			Name:     portName,
		})
		// to be filled in with applied ingress gateway tls port when building WorkloadEntries below
		workloadEntryPortMapping[portName] = 0
	}

	var workloadEntries []*networkingv1alpha3spec.WorkloadEntry

	// construct a WorkloadEntry for each ingress gateway destination's external address, for each endpoint (i.e. backing Workload) on the Destination
	for _, appliedIngressGateway := range destinationMesh.Status.AppliedEastWestIngressGateways {

		for portName := range workloadEntryPortMapping {
			workloadEntryPortMapping[portName] = appliedIngressGateway.DestinationPort
		}

		for _, externalAddress := range appliedIngressGateway.ExternalAddresses {
			for _, endpointSubset := range kubeService.EndpointSubsets {
				for _, endpoint := range endpointSubset.Endpoints {
					workloadEntry := &networkingv1alpha3spec.WorkloadEntry{
						Address: externalAddress,
						Ports:   workloadEntryPortMapping,
						Labels:  endpoint.Labels,
					}
					workloadEntries = append(workloadEntries, workloadEntry)
				}
			}
		}
	}

	federatedHostname := destination.Status.AppliedFederation.GetFederatedHostname()

	resolution, err := ResolutionForEndpointIpVersions(workloadEntries)
	if err != nil {
		return nil, err
	}

	// ObjectMeta's Namespace and ClusterName will be populated when translating federated outputs
	return &networkingv1alpha3.ServiceEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:   federatedHostname,
			Labels: metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.ServiceEntry{
			Addresses:  []string{serviceEntryIP.String()},
			Hosts:      []string{federatedHostname},
			Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
			Resolution: resolution,
			Endpoints:  workloadEntries,
			Ports:      sePorts,
		},
	}, nil
}

// translate resources local to this Destination's mesh that allow routing to this Destination from clients in remote Meshes
// A ServiceEntry is needed to map the Destination's global FQDN to the local FQDN.
func (t *translator) translateForLocalMesh(
	destination *discoveryv1.Destination,
	destinationMesh *discoveryv1.Mesh,
	remoteServiceEntryTemplate *networkingv1alpha3.ServiceEntry,
	remoteDestinationRule *networkingv1alpha3.DestinationRule,
) (*networkingv1alpha3.ServiceEntry, *networkingv1alpha3.DestinationRule, error) {
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

	resolution, err := ResolutionForEndpointIpVersions(workloadEntries)
	if err != nil {
		return nil, nil, err
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
			Resolution: resolution,
			Endpoints:  workloadEntries,
			Ports:      remoteServiceEntryTemplate.Spec.Ports,
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

	return se, dr, nil
}

// translate resources for remote meshes that allow routing to this Destination from clients in those remote Meshes
// A ServiceEntry is needed to represent the federated Destination on all remote meshes.
// A VirtualService and DestinationRule are needed to reflect any policies that apply to the federated Destination.
func (t *translator) translateForRemoteMesh(
	destination *discoveryv1.Destination,
	destinationVirtualMesh *networkingv1.VirtualMesh,
	in input.LocalSnapshot,
	remoteMesh *discoveryv1.Mesh,
	serviceEntryTemplate *networkingv1alpha3.ServiceEntry,
	reporter reporting.Reporter,
) (
	*networkingv1alpha3.ServiceEntry,
	*networkingv1alpha3.VirtualService,
	*networkingv1alpha3.DestinationRule,
) {
	remoteIstioMesh := remoteMesh.Spec.GetIstio()

	if getHostnameSuffix(destination.Status.AppliedFederation.GetFederatedHostname()) != hostutils.DefaultHostnameSuffix && !remoteIstioMesh.SmartDnsProxyingEnabled {
		reporter.ReportVirtualMeshToDestination(destination, destinationVirtualMesh, eris.Errorf(
			"mesh %v does not have smart DNS proxying enabled (hostname suffix can only be specified for federated Destination if Istio's smart DNS proxying is enabled)",
			sets.Key(remoteMesh),
		))
		return nil, nil, nil
	}

	// translate ServiceEntry
	serviceEntry := serviceEntryTemplate.DeepCopy()
	// set the Namespace and ClusterName based on the remote istio Mesh
	serviceEntry.Namespace = remoteIstioMesh.Installation.Namespace
	serviceEntry.ClusterName = remoteIstioMesh.Installation.Cluster

	// translate VirtualService for federated Destinations, can be nil
	virtualService := t.virtualServiceTranslator.Translate(t.ctx, in, destination, remoteIstioMesh.Installation, reporter)

	// translate DestinationRule for federated Destinations, can be nil
	destinationRule := t.destinationRuleTranslator.Translate(t.ctx, in, destination, remoteIstioMesh.Installation, reporter)

	return serviceEntry, virtualService, destinationRule
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

func getHostnameSuffix(hostname string) string {
	split := strings.Split(hostname, ".")
	return split[len(split)-1]
}

// Workaround for Istio issue where ipv6 addresses are supplied to Envoy incorrectly: https://github.com/envoyproxy/envoy/issues/10489#issuecomment-606290733.
// If any endpoints have an ipv6 address, set the resolution to STATIC,
// else, set the resolution to DNS.
// exported for use in enterprise
func ResolutionForEndpointIpVersions(
	workloadEntries []*networkingv1alpha3spec.WorkloadEntry,
) (networkingv1alpha3spec.ServiceEntry_Resolution, error) {
	var foundHostname bool
	var foundIpv6 bool
	for _, workloadEntry := range workloadEntries {
		// only ipv6 addresses will have 2 or more colons, reference: https://datatracker.ietf.org/doc/html/rfc5952
		if strings.Count(workloadEntry.Address, ":") >= 2 {
			foundIpv6 = true
		}
		if ip := net.ParseIP(workloadEntry.Address); ip == nil {
			foundHostname = true
		}
	}

	if foundHostname && foundIpv6 {
		return networkingv1alpha3spec.ServiceEntry_NONE, eris.New("endpoints contain both ipv6 and hostname addresses")
	} else if foundHostname && !foundIpv6 {
		return networkingv1alpha3spec.ServiceEntry_DNS, nil
	}
	return networkingv1alpha3spec.ServiceEntry_STATIC, nil
}
