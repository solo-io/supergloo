package istio

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/utils"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	istiov1alpha1 "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/util/gogoprotomarshal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	legacyPilotDeploymentName = "istio-pilot"
	istiodDeploymentName      = "istiod"
	istioContainerKeyword     = "istio"
	pilotContainerKeyword     = "pilot"
	istioConfigMapName        = "istio"
	istioConfigMapMeshDataKey = "mesh"

	// https://istio.io/docs/ops/deployment/requirements/#ports-used-by-istio
	defaultGatewayPortName = "tls"
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
func (d *meshDetector) DetectMeshes(in input.Snapshot) (v1alpha2.MeshSlice, error) {
	var meshes v1alpha2.MeshSlice
	var errs error
	for _, deployment := range in.Deployments().List() {
		mesh, err := d.detectMesh(deployment, in)
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

func (d *meshDetector) detectMesh(deployment *appsv1.Deployment, in input.Snapshot) (*v1alpha2.Mesh, error) {
	version, err := d.getIstiodVersion(deployment)
	if err != nil {
		return nil, err
	}

	if version == "" {
		return nil, nil
	}

	trustDomain, err := getTrustDomain(in.ConfigMaps(), deployment.ClusterName, deployment.Namespace)
	if err != nil {
		return nil, err
	}

	// TODO(ilackarms): allow configuring ingress gateway workload labels
	ingressGateways := getIngressGateways(
		d.ctx,
		deployment.Namespace,
		deployment.ClusterName,
		defaults.DefaultGatewayWorkloadLabels,
		in.Services(),
		in.Pods(),
		in.Nodes(),
	)

	agent := getAgent(
		deployment.ClusterName,
		in.Pods(),
	)

	mesh := &v1alpha2.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(deployment),
		Spec: v1alpha2.MeshSpec{
			MeshType: &v1alpha2.MeshSpec_Istio_{
				Istio: &v1alpha2.MeshSpec_Istio{
					Installation: &v1alpha2.MeshSpec_MeshInstallation{
						Namespace: deployment.Namespace,
						Cluster:   deployment.ClusterName,
						PodLabels: deployment.Spec.Selector.MatchLabels,
						Version:   version,
					},
					CitadelInfo: &v1alpha2.MeshSpec_Istio_CitadelInfo{
						TrustDomain: trustDomain,
						// This assumes that the istiod deployment is the cert provider
						CitadelServiceAccount: deployment.Spec.Template.Spec.ServiceAccountName,
					},
					IngressGateways: ingressGateways,
				},
			},
			AgentInfo: agent,
		},
	}

	return mesh, nil
}

func getIngressGateways(
	ctx context.Context,
	namespace string,
	clusterName string,
	workloadLabels map[string]string,
	allServices corev1sets.ServiceSet,
	allPods corev1sets.PodSet,
	allNodes corev1sets.NodeSet,
) []*v1alpha2.MeshSpec_Istio_IngressGatewayInfo {
	ingressSvcs := getIngressServices(allServices, namespace, clusterName, workloadLabels)
	var ingressGateways []*v1alpha2.MeshSpec_Istio_IngressGatewayInfo
	for _, svc := range ingressSvcs {
		gateway, err := getIngressGateway(svc, workloadLabels, allPods, allNodes)
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
	allPods corev1sets.PodSet,
	allNodes corev1sets.NodeSet,
) (*v1alpha2.MeshSpec_Istio_IngressGatewayInfo, error) {
	var (
		tlsPort *corev1.ServicePort
	)
	for _, port := range svc.Spec.Ports {
		port := port // pike
		if port.Name == defaultGatewayPortName {
			tlsPort = &port
			break
		}
	}
	if tlsPort == nil {
		return nil, eris.Errorf("no TLS port found on ingress gateway")
	}

	var (
		// TODO(ilackarms): currently we only use one address to connect to the gateway.
		// We can support multiple addresses per gateway for load balancing purposes in the future

		externalAddress string
		externalPort    uint32
	)
	switch svc.Spec.Type {
	case corev1.ServiceTypeNodePort:
		addr, err := getNodeIp(
			svc.ClusterName,
			svc.Namespace,
			workloadLabels,
			allPods,
			allNodes,
		)
		if err != nil {
			return nil, err
		}
		externalAddress = addr
		externalPort = uint32(tlsPort.NodePort)

	case corev1.ServiceTypeLoadBalancer:
		ingress := svc.Status.LoadBalancer.Ingress
		if len(ingress) == 0 {
			return nil, eris.Errorf("no loadBalancer.ingress status reported for service")
		}

		externalAddress = ingress[0].Hostname
		if externalAddress == "" {
			externalAddress = ingress[0].IP
		}
		externalPort = uint32(tlsPort.Port)

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

	return &v1alpha2.MeshSpec_Istio_IngressGatewayInfo{
		WorkloadLabels:   workloadLabels,
		ExternalAddress:  externalAddress,
		ExternalTlsPort:  externalPort,
		TlsContainerPort: uint32(containerPort),
	}, nil
}

func getIngressServices(
	allServices corev1sets.ServiceSet,
	namespace string,
	clusterName string,
	workloadLabels map[string]string,
) []*corev1.Service {
	return allServices.List(func(svc *corev1.Service) bool {
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
	return (deployment.GetName() == istiodDeploymentName || deployment.GetName() == legacyPilotDeploymentName) &&
		strings.Contains(container.Image, istioContainerKeyword) &&
		strings.Contains(container.Image, pilotContainerKeyword)
}

func getTrustDomain(
	configMaps corev1sets.ConfigMapSet,
	cluster,
	namespace string,
) (string, error) {
	istioConfigMap, err := configMaps.Find(&skv1.ClusterObjectRef{
		Name:        istioConfigMapName,
		Namespace:   namespace,
		ClusterName: cluster,
	})
	if err != nil {
		return "", err
	}

	meshConfigString, ok := istioConfigMap.Data[istioConfigMapMeshDataKey]
	if !ok {
		return "", eris.Errorf("Failed to find 'mesh' entry in ConfigMap with name/namespace/cluster %s/%s/%s", istioConfigMapName, namespace, cluster)
	}
	var meshConfig istiov1alpha1.MeshConfig
	err = gogoprotomarshal.ApplyYAML(meshConfigString, &meshConfig)
	if err != nil {
		return "", eris.Errorf("Failed to find 'mesh' entry in ConfigMap with name/namespace/cluster %s/%s/%s", istioConfigMapName, namespace, cluster)
	}
	return meshConfig.TrustDomain, nil
}

type Agent struct {
	Namespace string
}

func getAgent(
	cluster string,
	pods corev1sets.PodSet,
) *v1alpha2.MeshSpec_AgentInfo {
	agentPods := pods.List(func(pod *corev1.Pod) bool {
		return pod.ClusterName != cluster ||
			!labels.SelectorFromSet(agentLabels).Matches(labels.Set(pod.Labels))
	})
	if len(agentPods) == 0 {
		return nil
	}
	// currently assume only one agent installed per cluster/mesh
	agentNamespace := agentPods[0].Namespace
	return &v1alpha2.MeshSpec_AgentInfo{
		AgentNamespace: agentNamespace,
	}
}
