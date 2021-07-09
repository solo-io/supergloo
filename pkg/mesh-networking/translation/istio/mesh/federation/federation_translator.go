package federation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
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
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./federation_translator.go -destination mocks/federation_translator.go

const (
	// NOTE(ilackarms): we may want to support federating over non-tls port at some point.
	defaultGatewayProtocol = "TLS"
)

// the Mesh Federation translator translates a Gateway CR for enabling the Mesh to receive cross cluster traffic
type Translator interface {
	// Translate translates a Gateway for the given Mesh
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

	if len(mesh.Status.GetAppliedEastWestIngressGateways()) == 0 {
		reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.GetRef(),
			eris.Errorf("no Destinations selected as ingress gateway for mesh %v. At least one must be selected.", sets.Key(mesh)))
		return
	}

	// translate one Gateway CR per ingress gateway Destination
	for _, appliedIngressGateway := range mesh.Status.GetAppliedEastWestIngressGateways() {
		destination, err := in.Destinations().Find(ezkube.MakeObjectRef(appliedIngressGateway.GetDestinationRef()))
		if err != nil {
			contextutils.LoggerFrom(t.ctx).DPanicf("internal error: applied east west ingress gateway Destination not found in snapshot")
			continue
		}

		ingressContainerPort := appliedIngressGateway.GetContainerPort()
		if ingressContainerPort == 0 {
			contextutils.LoggerFrom(t.ctx).DPanicf("ingress gateway tls container port not found")
			continue
		}

		gateway := t.buildGateway(
			BuildGatewayName(appliedIngressGateway),
			istioNamespace,
			istioCluster,
			ingressContainerPort,
			federatedHostnameSuffix,
			destination.Spec.GetKubeService().GetWorkloadSelectorLabels(),
			virtualMesh.GetRef(),
		)

		outputs.AddGateways(gateway)
	}
}

func (t *translator) buildGateway(
	name, namespace, cluster string,
	ingressContainerPort uint32,
	federatedHostnameSuffix string,
	ingressGatewayWorkloadLabels map[string]string,
	virtualMeshRef ezkube.ResourceId,
) *networkingv1alpha3.Gateway {
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
					Number:   ingressContainerPort,
					Protocol: defaultGatewayProtocol,
					Name:     defaults.IstioGatewayTlsPortName,
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

	return gw
}

func BuildGatewayName(appliedIngressGateway *discoveryv1.MeshStatus_AppliedIngressGateway) string {
	ingressDestinationRef := appliedIngressGateway.GetDestinationRef()
	return kubeutils.SanitizeNameV2(
		fmt.Sprintf("%s-%s", ingressDestinationRef.GetName(), ingressDestinationRef.GetNamespace()),
	)
}
