package istio

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
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
	legacyPilotDeploymentName = "istio-pilot"
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
		ctx: contextutils.WithLogger(ctx, "mesh-detector"),
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

	ingressGateways := getIngressGateways(
		d.ctx,
		deployment.ClusterName,
		ingressGatewayDetector,
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
		contextutils.LoggerFrom(d.ctx).Warnw("could not get region for cluster", deployment.ClusterName, zap.Error(err))
	}

	mesh := &discoveryv1.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(deployment),
		Spec: discoveryv1.MeshSpec{
			Type: &discoveryv1.MeshSpec_Istio_{
				Istio: &discoveryv1.MeshSpec_Istio{
					Installation: &discoveryv1.MeshSpec_MeshInstallation{
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

// discover ingress gateway workload and destination metadata using label sets specified in settings, across all namespaces
func getIngressGateways(
	ctx context.Context,
	clusterName string,
	ingressGatewayDetector *settingsv1.DiscoverySettings_Istio_IngressGatewayDetector,
	services corev1sets.ServiceSet,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
) []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo {

	ingressSvcs := getIngressServices(services, clusterName, ingressGatewayDetector)
	var ingressGateways []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo
	for _, svc := range ingressSvcs {
		gateway, err := getIngressGateway(ctx, svc, ingressGatewayDetector.GatewayTlsPortName, pods, nodes)
		if err != nil {
			contextutils.LoggerFrom(ctx).Warnw("detection failed for matching istio ingress service", "error", err, "service", sets.Key(svc))
			continue
		}
		ingressGateways = append(ingressGateways, gateway)
	}
	return ingressGateways
}

func getIngressGateway(
	ctx context.Context,
	svc *corev1.Service,
	gatewayTlsPortName string,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
) (*discoveryv1.MeshSpec_Istio_IngressGatewayInfo, error) {
	var (
		tlsPort *corev1.ServicePort
	)
	for _, port := range svc.Spec.Ports {
		port := port // pike
		if port.Name == gatewayTlsPortName {
			tlsPort = &port
			break
		}
	}
	if tlsPort == nil {
		return nil, eris.Errorf("no TLS port found on ingress gateway")
	}

	if tlsPort.TargetPort.StrVal != "" {
		// TODO(ilackarms): for the sake of simplicity, we only support number target ports
		// if we come across the need to support named ports, we can add the lookup on the pod container itself here
		return nil, eris.Errorf("named target ports are not currently supported on ingress gateway")
	}

	var externalAddresses []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress
	var externalTlsPort uint32
	var labelSets []*commonv1.LabelSet
	switch svc.Spec.Type {
	case corev1.ServiceTypeNodePort:
		var nodeIps []string
		labelSets, nodeIps = getWorkloadLabelSetsAndNodeIps(
			ctx,
			svc.ClusterName,
			svc.Namespace,
			svc.Spec.Selector,
			pods,
			nodes,
		)
		for _, nodeIp := range nodeIps {
			externalAddresses = append(externalAddresses, &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress{
				ExternalAddressType: &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress_Ip{
					Ip: nodeIp,
				},
			})
		}
		externalTlsPort = uint32(tlsPort.NodePort)
	case corev1.ServiceTypeLoadBalancer:
		ingress := svc.Status.LoadBalancer.Ingress
		for _, loadBalancerIngress := range ingress {

			var externalAddress *discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress
			// If the Ip address is set in the ingress, use that
			if loadBalancerIngress.IP != "" {
				externalAddress = &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress{
					ExternalAddressType: &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress_Ip{
						Ip: loadBalancerIngress.IP,
					},
				}
			} else if loadBalancerIngress.Hostname != "" {
				// Otherwise use the hostname
				externalAddress = &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress{
					ExternalAddressType: &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress_DnsName{
						DnsName: loadBalancerIngress.Hostname,
					},
				}
			}
			externalAddresses = append(externalAddresses, externalAddress)
		}
		externalTlsPort = uint32(tlsPort.Port)
	default:
		return nil, eris.Errorf("unsupported service type %v for ingress gateway", svc.Spec.Type)
	}

	// TODO: distinguish between these manually assigned external IPs and assigned IPs in case a user may want to specify which IPs to use for networking policies
	for _, externalIP := range svc.Spec.ExternalIPs {
		externalAddresses = append(externalAddresses, &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress{
			ExternalAddressType: &discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress_Ip{
				Ip: externalIP,
			},
		})
	}

	containerPort := tlsPort.TargetPort.IntVal
	if containerPort == 0 {
		containerPort = tlsPort.Port
	}

	if len(externalAddresses) == 0 {
		return nil, eris.Errorf("no external addresses found for service type %v for ingress gateway", svc.Spec.Type)
	}

	// set deprecated field using arbitrary first label set
	var workloadLabels map[string]string
	if len(labelSets) > 0 {
		workloadLabels = labelSets[0].Labels
	}

	gatewayInfo := &discoveryv1.MeshSpec_Istio_IngressGatewayInfo{
		WorkloadLabels:    workloadLabels,
		WorkloadLabelSets: labelSets,
		ExternalTlsPort:   externalTlsPort,
		TlsContainerPort:  uint32(containerPort),
		ExternalAddresses: externalAddresses,
	}

	// Continue to set deprecated fields until they are removed, only populate with first external address.
	switch extAddr := externalAddresses[0].ExternalAddressType.(type) {
	case *discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress_Ip:
		gatewayInfo.ExternalAddress = extAddr.Ip
	case *discoveryv1.MeshSpec_Istio_IngressGatewayInfo_ExternalAddress_DnsName:
		gatewayInfo.ExternalAddress = extAddr.DnsName
	}

	return gatewayInfo, nil
}

func getIngressServices(
	services corev1sets.ServiceSet,
	clusterName string,
	ingressGatewayDetector *settingsv1.DiscoverySettings_Istio_IngressGatewayDetector,
) []*corev1.Service {
	var ingressServices []*corev1.Service
	services.List(func(svc *corev1.Service) (_ bool) {
		for _, workloadLabels := range ingressGatewayDetector.GatewayWorkloadLabelSets {
			if svc.ClusterName == clusterName && labels.SelectorFromSet(workloadLabels.Labels).Matches(labels.Set(svc.Spec.Selector)) {
				ingressServices = append(ingressServices, svc)
				break
			}
		}
		return
	})

	return ingressServices
}

func getWorkloadLabelSetsAndNodeIps(
	ctx context.Context,
	svcCluster,
	svcNamespace string,
	svcSelector map[string]string,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
) ([]*commonv1.LabelSet, []string) {
	var labelSets []*commonv1.LabelSet
	var ips []string

	var ingressPods []*corev1.Pod
	pods.List(func(pod *corev1.Pod) (_ bool) {
		if pod.ClusterName == svcCluster &&
			pod.Namespace == svcNamespace &&
			labels.SelectorFromSet(svcSelector).Matches(labels.Set(pod.Labels)) {

			ingressPods = append(ingressPods, pod)
			labelSets = append(labelSets, &commonv1.LabelSet{Labels: pod.Labels})
		}
		return
	})

	if len(ingressPods) < 1 {
		contextutils.LoggerFrom(ctx).Warnf("no backing pods found for ingress service selector %v in namespace %v", svcSelector, svcNamespace)
		return nil, nil
	}
	// TODO(ilackarms): currently we just grab the node ip of the first available pod
	// Eventually we may want to consider supporting multiple nodes/IPs for load balancing.
	var ingressNodeNames []string
	for _, ingressPod := range ingressPods {
		ingressNode, err := nodes.Find(&skv1.ClusterObjectRef{
			ClusterName: svcCluster,
			Name:        ingressPod.Spec.NodeName,
		})
		if err != nil {
			contextutils.LoggerFrom(ctx).DPanicf("internal error: failed to find ingress node for pod %v", sets.Key(ingressPod))
			return nil, nil
		}
		ingressNodeNames = append(ingressNodeNames, ingressPod.Spec.NodeName)

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
			ips = append(ips, addr.Address)
		}
	}
	if len(ips) == 0 {
		contextutils.LoggerFrom(ctx).Warnf("no external IP addresses assigned for ingress node %v", ingressNodeNames)
		return nil, nil
	}

	return labelSets, ips
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
	return (deployment.GetName() == istiodDeploymentName || deployment.GetName() == legacyPilotDeploymentName) &&
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
