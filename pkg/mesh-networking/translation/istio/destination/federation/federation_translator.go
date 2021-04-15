package federation

import (
	"context"

	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/trafficshift"
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

	destinationVirtualMesh, err := in.VirtualMeshes().Find(destination.Status.AppliedFederation.GetRef())
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("Could not find parent VirtualMesh %v for Destination %v", destination.Status.AppliedFederation.GetRef(), ezkube.MakeObjectRef(destination))
		return nil, nil, nil
	}

	// translate ServiceEntry template
	serviceEntryTemplate, err := t.translateServiceEntryTemplate(destination, destinationMesh)
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("Encountered error while translating ServiceEntry template for Destination %v", ezkube.MakeObjectRef(destination))
		return nil, nil, nil
	}

	var serviceEntries []*networkingv1alpha3.ServiceEntry
	var virtualServices []*networkingv1alpha3.VirtualService
	var destinationRules []*networkingv1alpha3.DestinationRule

	for _, meshRef := range destination.Status.AppliedFederation.GetFederatedToMeshes() {
		remoteMesh, err := in.Meshes().Find(meshRef)
		if err != nil {
			contextutils.LoggerFrom(t.ctx).Errorf("Could not find Mesh %v that Destination %v is federated to", meshRef, ezkube.MakeObjectRef(destination))
			continue
		}

		serviceEntry, virtualService, destinationRule := t.translateForRemoteMesh(
			destination,
			destinationMesh,
			destinationVirtualMesh,
			in,
			remoteMesh,
			serviceEntryTemplate,
			reporter,
		)

		// Append the VirtualMesh as a parent to the outputs
		// TODO(harveyxia) append Destination as parent once new GC is being implemented
		metautils.AppendParent(t.ctx, serviceEntry, destination.Status.AppliedFederation.GetRef(), networkingv1.VirtualMesh{}.GVK())
		metautils.AppendParent(t.ctx, virtualService, destination.Status.AppliedFederation.GetRef(), networkingv1.VirtualMesh{}.GVK())
		metautils.AppendParent(t.ctx, destinationRule, destination.Status.AppliedFederation.GetRef(), networkingv1.VirtualMesh{}.GVK())

		serviceEntries = append(serviceEntries, serviceEntry)
		virtualServices = append(virtualServices, virtualService)
		destinationRules = append(destinationRules, destinationRule)
	}

	return serviceEntries, virtualServices, destinationRules
}

// translate the ServiceEntry template that must exist on all meshes this Destination is federated to
func (t *translator) translateServiceEntryTemplate(
	destination *discoveryv1.Destination,
	destinationMesh *discoveryv1.Mesh,
) (*networkingv1alpha3.ServiceEntry, error) {
	istioCluster := destinationMesh.Spec.GetIstio().Installation.Cluster

	// Guaranteed to have at least one gateway passed by caller
	ingressGateway := destinationMesh.Spec.GetIstio().IngressGateways[0]
	serviceEntryIP, err := destinationutils.ConstructUniqueIpForKubeService(destination.Spec.GetKubeService().GetRef())
	if err != nil {
		// should never happen
		return nil, eris.Errorf("unexpected error: failed to generate service entry ip: %v", err)
	}

	endpointPorts := make(map[string]uint32)
	var ports []*networkingv1alpha3spec.Port
	for _, port := range destination.Spec.GetKubeService().GetPorts() {
		ports = append(ports, &networkingv1alpha3spec.Port{
			Number:   port.Port,
			Protocol: ConvertKubePortProtocol(port),
			Name:     port.Name,
		})
		endpointPorts[port.Name] = ingressGateway.ExternalTlsPort
	}

	// NOTE(ilackarms): we use these labels to support federated subsets.
	// the values don't actually matter; but the subset names should
	// match those on the DestinationRule for the Destination in the
	// remote cluster.
	// based on: https://istio.io/latest/blog/2019/multicluster-version-routing/#create-a-destination-rule-on-both-clusters-for-the-local-reviews-service
	clusterLabels := trafficshift.MakeFederatedSubsetLabel(istioCluster)

	address := ingressGateway.GetDnsName()
	if address == "" {
		address = ingressGateway.GetIp()
	}
	endpoints := []*networkingv1alpha3spec.WorkloadEntry{{
		Address: address,
		Ports:   endpointPorts,
		Labels:  clusterLabels,
	}}

	federatedHostname := destination.Status.AppliedFederation.GetFederatedHostname()

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
			Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
			Endpoints:  endpoints,
			Ports:      ports,
		},
	}, nil
}

func (t *translator) translateForRemoteMesh(
	destination *discoveryv1.Destination,
	destinationMesh *discoveryv1.Mesh,
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

	if destination.Status.AppliedFederation.GetFederatedHostname() != hostutils.DefaultHostnameSuffix && !remoteIstioMesh.SmartDnsProxyingEnabled {
		reporter.ReportVirtualMeshToMesh(destinationMesh, destinationVirtualMesh, eris.Errorf(
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
