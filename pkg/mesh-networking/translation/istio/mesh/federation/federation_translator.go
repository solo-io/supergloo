package federation

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/gogo/protobuf/types"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/protoutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/envoy/config/filter/network/tcp_cluster_rewrite/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./federation_translator.go -destination mocks/federation_translator.go

const (
	// NOTE(ilackarms): we may want to support federating over non-tls port at some point.
	defaultGatewayProtocol = "TLS"
	DefaultGatewayPortName = "tls"

	envoySniClusterFilterName        = "envoy.envoyfilters.network.sni_cluster"
	envoyTcpClusterRewriteFilterName = "envoy.envoyfilters.network.tcp_cluster_rewrite"
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

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	in input.LocalSnapshot,
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	outputs istio.Builder,
	_ reporting.Reporter,
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

	istioCluster := istioMesh.Installation.Cluster

	kubeCluster, err := in.KubernetesClusters().Find(&skv2corev1.ObjectRef{
		Name:      istioCluster,
		Namespace: defaults.GetPodNamespace(),
	})
	if err != nil {
		contextutils.LoggerFrom(t.ctx).Errorf("internal error: cluster %v for istio mesh %v not found", istioCluster, sets.Key(mesh))
		return
	}

	istioNamespace := istioMesh.Installation.Namespace

	federatedHostnameSuffix := hostutils.GetFederatedHostnameSuffix(virtualMesh.Spec)

	tcpRewritePatch, err := buildTcpRewritePatch(
		istioMesh,
		istioCluster,
		kubeCluster.Spec.ClusterDomain,
		federatedHostnameSuffix,
	)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(t.ctx).DPanicf("failed generating tcp rewrite patch: %v", err)
		return
	}

	// istio gateway names must be DNS-1123 labels
	// hyphens are legal, dots are not, so we convert here
	gwName := BuildGatewayName(virtualMesh)
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

	ef := BuildTcpRewriteEnvoyFilter(
		ingressGateway,
		istioMesh.GetInstallation(),
		tcpRewritePatch,
		fmt.Sprintf("%v.%v", virtualMesh.Ref.Name, virtualMesh.Ref.Namespace),
	)

	// Append the virtual mesh as a parent to each output resource
	metautils.AppendParent(t.ctx, gw, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())
	metautils.AppendParent(t.ctx, ef, virtualMesh.GetRef(), networkingv1.VirtualMesh{}.GVK())

	outputs.AddGateways(gw)
	outputs.AddEnvoyFilters(ef)
}

func BuildGatewayName(virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh) string {
	return kubeutils.SanitizeNameV2(
		fmt.Sprintf("%s-%s", virtualMesh.GetRef().GetName(), virtualMesh.GetRef().GetNamespace()),
	)
}

func BuildTcpRewriteEnvoyFilter(
	ingressGateway *discoveryv1.MeshSpec_Istio_IngressGatewayInfo,
	istioInstallation *discoveryv1.MeshSpec_MeshInstallation,
	tcpRewritePatch *types.Struct,
	name string,
) *networkingv1alpha3.EnvoyFilter {
	return &networkingv1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   istioInstallation.GetNamespace(),
			ClusterName: istioInstallation.GetCluster(),
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
}

func buildTcpRewritePatch(
	istioMesh *discoveryv1.MeshSpec_Istio,
	clusterName string,
	clusterDomain string,
	federatedHostnameSuffix string,
) (*types.Struct, error) {
	if clusterDomain == "" {
		clusterDomain = defaults.DefaultClusterDomain
	}
	clusterPattern := fmt.Sprintf("\\.%s.%s$", clusterName, federatedHostnameSuffix)
	clusterReplacement := "." + clusterDomain
	return BuildTcpRewritePatch(istioMesh, clusterPattern, clusterReplacement)
}

// BuildTcpRewritePatch Public to be used in enterprise
func BuildTcpRewritePatch(
	istioMesh *discoveryv1.MeshSpec_Istio,
	clusterPattern, clusterReplacement string,
) (*types.Struct, error) {
	version, err := semver.NewVersion(istioMesh.Installation.Version)
	if err != nil {
		return nil, err
	}
	constraint, err := semver.NewConstraint("<= 1.6.8")
	if err != nil {
		return nil, err
	}
	// If Istio version less than 1.7.x, use untyped config
	if constraint.Check(version) {
		return buildTcpRewritePatchAsConfig(clusterPattern, clusterReplacement)

	}
	// If Istio version >= 1.7.x, used typed config
	return buildTcpRewritePatchAsTypedConfig(clusterPattern, clusterReplacement)
}

func buildTcpRewritePatchAsTypedConfig(clusterPattern, clusterReplacement string) (*types.Struct, error) {
	tcpClusterRewrite, err := protoutils.MessageToAnyWithError(&v2alpha1.TcpClusterRewrite{
		ClusterPattern:     clusterPattern,
		ClusterReplacement: clusterReplacement,
	})
	if err != nil {
		return nil, err
	}
	return protoutils.GolangMessageToGogoStruct(&envoy_config_listener_v3.Filter{
		Name: envoyTcpClusterRewriteFilterName,
		ConfigType: &envoy_config_listener_v3.Filter_TypedConfig{
			TypedConfig: tcpClusterRewrite,
		},
	})
}

func buildTcpRewritePatchAsConfig(clusterPattern, clusterReplacement string) (*types.Struct, error) {
	tcpRewrite, err := protoutils.GogoMessageToGolangStruct(&v2alpha1.TcpClusterRewrite{
		ClusterPattern:     clusterPattern,
		ClusterReplacement: clusterReplacement,
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
