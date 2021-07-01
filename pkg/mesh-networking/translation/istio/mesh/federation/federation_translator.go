package federation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"

	"github.com/solo-io/skv2/pkg/ezkube"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./federation_translator.go -destination mocks/federation_translator.go

const (
	// NOTE(ilackarms): we may want to support federating over non-tls port at some point.
	defaultGatewayProtocol = "TLS"
	defaultGatewayPort     = 15443
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
	ctx context.Context
}

func NewTranslator(
	ctx context.Context,
) Translator {
	return &translator{ctx: ctx}
}

type ingressGatewayInfo struct {
	labels map[string]string
	port   uint32
}

// translate the appropriate resources for the given Mesh.
// A Gateway is needed to configure the ingress gateway workload to forward requests originating from external meshes.
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

	istioCluster := istioMesh.Installation.Cluster
	istioNamespace := istioMesh.Installation.Namespace
	federatedHostnameSuffix := hostutils.GetFederatedHostnameSuffix(virtualMesh.Spec)

	if len(mesh.Status.GetEastWestIngressGateways()) == 0 {
		reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.GetRef(),
			eris.Errorf("No Destinations selected as ingress gateway for mesh %v. At least one must be selected.", sets.Key(mesh)))
		return
	}

	// create the east west ingress gateways
	for _, ewIngressGatewayInfo := range mesh.Status.GetEastWestIngressGateways() {
		destination, err := in.Destinations().Find(ewIngressGatewayInfo.GetDestinationRef())
		if err != nil {
			continue // should log an error in the VirtualMesh status
		}

		// Figure out the ingress gateway tls port
		ingressGatewayContainerTlsPort := uint32(defaultGatewayPort)
		gatewayTlsPortName := defaults.DefaultGatewayPortName
		if ewIngressGatewayInfo.GetTlsPortName() != "" {
			gatewayTlsPortName = ewIngressGatewayInfo.GetTlsPortName()
		}
		for _, ports := range destination.Spec.GetKubeService().GetPorts() {
			if ports.GetName() == gatewayTlsPortName {
				ingressGatewayContainerTlsPort = ports.GetPort()
				break
			}
		}

		// figure out the ingress gateway workload labels
		ingressGatewayWorkloadLabels := defaults.DefaultGatewayWorkloadLabels
		if len(destination.Spec.GetKubeService().GetWorkloadSelectorLabels()) != 0 {
			ingressGatewayWorkloadLabels = destination.Spec.GetKubeService().GetWorkloadSelectorLabels()
		}

		// istio gateway names must be DNS-1123 labels
		// hyphens are legal, dots are not, so we convert here
		gwName := BuildGatewayName(destination.GetName(), destination.GetNamespace())
		t.buildGatewayObject(gwName, istioNamespace, istioCluster,
			ingressGatewayContainerTlsPort, federatedHostnameSuffix, ingressGatewayWorkloadLabels,
			virtualMesh.GetRef(), outputs)
	}
}

func (t *translator) buildGatewayObject(name, namespace, cluster string, ingressGatewayContainerTlsPort uint32,
	federatedHostnameSuffix string, ingressGatewayWorkloadLabels map[string]string,
	virtualMeshRef ezkube.ResourceId, outputs istio.Builder) {
	gw := &networkingv1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			ClusterName: cluster,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: networkingv1alpha3spec.Gateway{
			Servers: []*networkingv1alpha3spec.Server{{
				Port: &networkingv1alpha3spec.Port{
					Number:   ingressGatewayContainerTlsPort,
					Protocol: defaultGatewayProtocol,
					Name:     defaults.DefaultGatewayPortName,
				},
				Hosts: []string{"*." + federatedHostnameSuffix},
				Tls: &networkingv1alpha3spec.ServerTLSSettings{
					Mode: networkingv1alpha3spec.ServerTLSSettings_AUTO_PASSTHROUGH,
				},
			}},
			Selector: ingressGatewayWorkloadLabels,
		},
	}

	// Append the virtual mesh as a parent to each output resource
	metautils.AppendParent(t.ctx, gw, virtualMeshRef, networkingv1.VirtualMesh{}.GVK())
	outputs.AddGateways(gw)
}

func BuildGatewayName(name, namespace string) string {
	return kubeutils.SanitizeNameV2(
		fmt.Sprintf("%s-%s", name, namespace),
	)
}
