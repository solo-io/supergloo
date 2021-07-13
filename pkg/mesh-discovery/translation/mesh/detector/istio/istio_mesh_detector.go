package istio

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/localityutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"go.uber.org/zap"
	istiov1alpha1 "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/util/gogoprotomarshal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	istiodDeploymentName      = "istiod"
	istioContainerKeyword     = "istio"
	pilotContainerKeyword     = "pilot"
	istioConfigMapName        = "istio"
	istioConfigMapMeshDataKey = "mesh"
	istioMetaDnsCaptureKey    = "ISTIO_META_DNS_CAPTURE"
)

var (
	// these labels are hard-coded to match the labels used on
	// the cert-agent pod template in the cert-agent helm chart.
	agentLabels = map[string]string{"app": "cert-agent"}
)

// detects Istio if a deployment contains the istiod container.
type meshDetector struct {
	ctx context.Context
}

func NewMeshDetector(
	ctx context.Context,
) detector.MeshDetector {
	return &meshDetector{
		ctx: contextutils.WithLogger(ctx, "detector"),
	}
}

// returns a mesh for each deployment that contains the istiod image
func (d *meshDetector) DetectMeshes(
	in input.DiscoveryInputSnapshot,
	settings *settingsv1.DiscoverySettings,
) (discoveryv1.MeshSlice, error) {
	var meshes discoveryv1.MeshSlice
	var errs error
	for _, deployment := range in.Deployments().List() {
		mesh, err := d.detectMesh(deployment, in, settings)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		if mesh == nil {
			continue
		}
		meshes = append(meshes, mesh)
	}
	return meshes, errs
}

func (d *meshDetector) detectMesh(
	deployment *appsv1.Deployment,
	in input.DiscoveryInputSnapshot,
	settings *settingsv1.DiscoverySettings,
) (*discoveryv1.Mesh, error) {
	version, err := d.getIstiodVersion(deployment)
	if err != nil {
		return nil, err
	}

	if version == "" {
		return nil, nil
	}

	meshConfig, err := getMeshConfig(in.ConfigMaps(), deployment.ClusterName, deployment.Namespace)
	if err != nil {
		return nil, err
	}

	ingressGatewayDetector, err := utils.GetIngressGatewayDetector(settings, deployment.ClusterName)
	if err != nil {
		return nil, err
	}

	// TOOD (ilackarms / ingress): support discovering gateways deployed to namespace other than istio-system
	ingressGateways := getIngressGateways(
		d.ctx,
		deployment.Namespace,
		deployment.ClusterName,
		ingressGatewayDetector.GetGatewayWorkloadLabels(),
		ingressGatewayDetector.GetGatewayTlsPortName(),
		in.Services(),
		in.Pods(),
		in.Nodes(),
	)

	agent := getAgent(
		deployment.ClusterName,
		in.Pods(),
	)

	region, err := localityutils.GetClusterRegion(deployment.ClusterName, in.Nodes())
	if err != nil {
		contextutils.LoggerFrom(d.ctx).Debugw("could not get region for cluster", deployment.ClusterName, zap.Error(err))
	}

	mesh := &discoveryv1.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(deployment),
		Spec: discoveryv1.MeshSpec{
			Type: &discoveryv1.MeshSpec_Istio_{
				Istio: &discoveryv1.MeshSpec_Istio{
					Installation: &discoveryv1.MeshInstallation{
						Namespace: deployment.Namespace,
						Cluster:   deployment.ClusterName,
						PodLabels: deployment.Spec.Selector.MatchLabels,
						Version:   version,
						Region:    region,
					},
					SmartDnsProxyingEnabled: isSmartDnsProxyingEnabled(meshConfig),
					TrustDomain:             meshConfig.TrustDomain,
					// This assumes that the istiod deployment is the cert provider
					IstiodServiceAccount: deployment.Spec.Template.Spec.ServiceAccountName,
					IngressGateways:      ingressGateways,
				},
			},
			AgentInfo: agent,
		},
	}

	return mesh, nil
}

// DEPRECATED: in favor of Destination discovery, which captures all external address information
// TODO: remove this code when Mesh.spec.Type.Istio.IngressGateways is deleted
func getIngressGateways(
	ctx context.Context,
	namespace string,
	clusterName string,
	workloadLabels map[string]string,
	tlsPortName string,
	services corev1sets.ServiceSet,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
) []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo {
	ingressSvcs := getIngressServices(services, namespace, clusterName, workloadLabels)
	var ingressGateways []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo
	for _, svc := range ingressSvcs {
		gateway, err := getIngressGateway(svc, workloadLabels, tlsPortName, pods, nodes)
		if err != nil {
			contextutils.LoggerFrom(ctx).Warnw("detection failed for matching istio ingress service", "error", err, "service", sets.Key(svc))
			continue
		}
		ingressGateways = append(ingressGateways, gateway)
	}
	return ingressGateways
}

func getIngressGateway(
	svc *corev1.Service,
	workloadLabels map[string]string,
	tlsPortName string,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
) (*discoveryv1.MeshSpec_Istio_IngressGatewayInfo, error) {
	var (
		tlsPort *corev1.ServicePort
	)
	for _, port := range svc.Spec.Ports {
		port := port // pike
		if port.Name == tlsPortName {
			tlsPort = &port
			break
		}
	}
	if tlsPort == nil {
		return nil, eris.Errorf("no TLS port found on ingress gateway")
	}

	// TODO(ilackarms): currently we only use one address to connect to the gateway.
	// We can support multiple addresses per gateway for load balancing purposes in the future

	gatewayInfo := &discoveryv1.MeshSpec_Istio_IngressGatewayInfo{
		Namespace:      svc.Namespace,
		Name:           svc.Name,
		WorkloadLabels: svc.Spec.Selector,
	}

	switch svc.Spec.Type {
	case corev1.ServiceTypeNodePort:
		gatewayInfo.ExternalTlsPort = uint32(tlsPort.NodePort)
		addr, err := getNodeIp(
			svc.ClusterName,
			svc.Namespace,
			workloadLabels,
			pods,
			nodes,
		)
		if err != nil {
			// Check for user-set external IPs
			externalIPs := svc.Spec.ExternalIPs
			if len(externalIPs) != 0 {
				addr = svc.Spec.ExternalIPs[0]
			} else {
				return nil, err
			}
		}
		gatewayInfo.ExternalAddressType = &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_Ip{
			Ip: addr,
		}
		// Continue to set deprecated field until it is removed
		gatewayInfo.ExternalAddress = addr
	case corev1.ServiceTypeLoadBalancer:
		gatewayInfo.ExternalTlsPort = uint32(tlsPort.Port)
		ingress := svc.Status.LoadBalancer.Ingress
		var addr string
		if len(ingress) == 0 {
			// Check for user-set external IPs
			externalIPs := svc.Spec.ExternalIPs
			if len(externalIPs) != 0 {
				addr = svc.Spec.ExternalIPs[0]
				gatewayInfo.ExternalAddressType = &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_Ip{
					Ip: addr,
				}
			} else {
				return nil, eris.Errorf("no loadBalancer.ingress status reported for service. Please set an external IP on the service if you are using a non-kubernetes load balancer.")
			}
		} else if ingress[0].IP != "" {
			// If the Ip address is set in the ingress, use that
			addr = ingress[0].IP
			gatewayInfo.ExternalAddressType = &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_Ip{
				Ip: addr,
			}
		} else {
			// Otherwise use the hostname
			addr = ingress[0].Hostname
			gatewayInfo.ExternalAddressType = &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_DnsName{
				DnsName: addr,
			}
		}
		// Continue to set deprecated field until it is removed
		gatewayInfo.ExternalAddress = addr
	default:
		return nil, eris.Errorf("unsupported service type %v for ingress gateway", svc.Spec.Type)
	}

	if tlsPort.TargetPort.StrVal != "" {
		// TODO(ilackarms): for the sake of simplicity, we only support number target ports
		// if we come across the need to support named ports, we can add the lookup on the pod container itself here
		return nil, eris.Errorf("named target ports are not currently supported on ingress gateway")
	}
	containerPort := tlsPort.TargetPort.IntVal
	if containerPort == 0 {
		containerPort = tlsPort.Port
	}
	gatewayInfo.TlsContainerPort = uint32(containerPort)

	return gatewayInfo, nil
}

func getIngressServices(
	services corev1sets.ServiceSet,
	namespace string,
	clusterName string,
	workloadLabels map[string]string,
) []*corev1.Service {
	return services.List(func(svc *corev1.Service) bool {
		return svc.Namespace != namespace ||
			svc.ClusterName != clusterName ||
			!labels.SelectorFromSet(workloadLabels).Matches(labels.Set(svc.Spec.Selector))
	})
}

func getNodeIp(
	cluster,
	namespace string,
	workloadLabels map[string]string,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
) (string, error) {
	ingressPods := pods.List(func(pod *corev1.Pod) bool {
		return pod.ClusterName != cluster ||
			pod.Namespace != namespace ||
			!labels.SelectorFromSet(workloadLabels).Matches(labels.Set(pod.Labels))
	})
	if len(ingressPods) < 1 {
		return "", eris.Errorf("no pods found backing ingress workload %v in namespace %v", workloadLabels, namespace)
	}
	// TODO(ilackarms): currently we just grab the node ip of the first available pod
	// Eventually we may want to consider supporting multiple nodes/IPs for load balancing.
	ingressPod := ingressPods[0]
	ingressNode, err := nodes.Find(&skv1.ClusterObjectRef{
		ClusterName: cluster,
		Name:        ingressPod.Spec.NodeName,
	})
	if err != nil {
		return "", eris.Wrapf(err, "failed to find ingress node for pod %v", sets.Key(ingressPod))
	}

	isKindNode := isKindNode(ingressNode)
	for _, addr := range ingressNode.Status.Addresses {
		if isKindNode {
			// For Kind clusters, we use the NodeInteralIP for the external IP address.
			if addr.Type != corev1.NodeInternalIP {
				continue
			}
		} else if addr.Type == corev1.NodeInternalIP || addr.Type == corev1.NodeInternalDNS {
			continue
		}
		return addr.Address, nil
	}
	return "", eris.Errorf("no external addresses reported for ingress node %v", sets.Key(ingressNode))
}

func isKindNode(node *corev1.Node) bool {
	for _, image := range node.Status.Images {
		for _, name := range image.Names {
			if strings.Contains(name, "kindnetd") {
				return true
			}
		}
	}
	return false
}

func (d *meshDetector) getIstiodVersion(deployment *appsv1.Deployment) (string, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if isIstiod(deployment, &container) {
			parsedImage, err := dockerutils.ParseImageName(container.Image)
			if err != nil {
				return "", eris.Wrapf(err, "failed to parse istiod image tag: %s", container.Image)
			}
			version := parsedImage.Tag
			if parsedImage.Digest != "" {
				version = parsedImage.Digest
			}
			return version, nil
		}
	}
	return "", nil
}

// Return true if deployment is inferred to be an Istiod deployment
func isIstiod(deployment *appsv1.Deployment, container *corev1.Container) bool {
	// Istio revision deployments may take the form `istiod-<revision-name>`
	return strings.HasPrefix(deployment.GetName(), istiodDeploymentName) &&
		strings.Contains(container.Image, istioContainerKeyword) &&
		strings.Contains(container.Image, pilotContainerKeyword)
}

func getMeshConfig(
	configMaps corev1sets.ConfigMapSet,
	cluster,
	namespace string,
) (*istiov1alpha1.MeshConfig, error) {
	istioConfigMap, err := configMaps.Find(&skv1.ClusterObjectRef{
		Name:        istioConfigMapName,
		Namespace:   namespace,
		ClusterName: cluster,
	})
	if err != nil {
		return nil, err
	}

	meshConfigString, ok := istioConfigMap.Data[istioConfigMapMeshDataKey]
	if !ok {
		return nil, eris.Errorf("Failed to find 'mesh' entry in ConfigMap with name/namespace/cluster %s/%s/%s", istioConfigMapName, namespace, cluster)
	}
	var meshConfig istiov1alpha1.MeshConfig
	err = gogoprotomarshal.ApplyYAML(meshConfigString, &meshConfig)
	if err != nil {
		return nil, eris.Errorf("Failed to find 'mesh' entry in ConfigMap with name/namespace/cluster %s/%s/%s", istioConfigMapName, namespace, cluster)
	}
	return &meshConfig, nil
}

// Reference for Istio's "smart DNS proxying" feature, https://istio.io/latest/blog/2020/dns-proxy/
// Reference for ISTIO_META_DNS_CAPTURE env var: https://preliminary.istio.io/latest/docs/reference/commands/pilot-agent/#envvars
func isSmartDnsProxyingEnabled(meshConfig *istiov1alpha1.MeshConfig) bool {
	proxyMetadata := meshConfig.GetDefaultConfig().GetProxyMetadata()
	if proxyMetadata == nil {
		return false
	}
	return proxyMetadata[istioMetaDnsCaptureKey] == "true"
}

type Agent struct {
	Namespace string
}

func getAgent(
	cluster string,
	pods corev1sets.PodSet,
) *discoveryv1.MeshSpec_AgentInfo {
	agentNamespace := getCertAgentNamespace(cluster, pods)
	if agentNamespace == "" {
		return nil
	}
	return &discoveryv1.MeshSpec_AgentInfo{
		AgentNamespace: agentNamespace,
	}
}

func getCertAgentNamespace(
	cluster string,
	pods corev1sets.PodSet,
) string {
	if defaults.GetAgentCluster() != "" {
		// discovery is running as the agent, assume the cert agent runs in the same namespace
		return defaults.GetPodNamespace()
	}
	agentPods := pods.List(func(pod *corev1.Pod) bool {

		return pod.ClusterName != cluster ||
			!labels.SelectorFromSet(agentLabels).Matches(labels.Set(pod.Labels))
	})
	if len(agentPods) == 0 {
		return ""
	}
	// currently assume only one agent installed per cluster/mesh
	return agentPods[0].Namespace
}
