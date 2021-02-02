package istio_test

import (
	"context"
	"strconv"

	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	istiov1alpha1 "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/util/protomarshal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/mesh/detector/istio"
)

var _ = Describe("IstioMeshDetector", func() {
	ctx := context.TODO()
	serviceAccountName := "service-account-name"
	meshNs := "namespace"
	clusterName := "cluster"
	pilotDeploymentName := "istio-pilot"
	istiodDeploymentName := "istiod"

	istioDeployment := func(deploymentName string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   meshNs,
				Name:        deploymentName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "istio-pilot:latest",
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "istiod"},
				},
			},
		}
	}

	trustDomain := "cluster.local"
	smartDnsProxyingEnabled := true
	istioConfigMap := func() corev1sets.ConfigMapSet {
		meshConfig := &istiov1alpha1.MeshConfig{
			DefaultConfig: &istiov1alpha1.ProxyConfig{
				ProxyMetadata: map[string]string{
					"ISTIO_META_DNS_CAPTURE": strconv.FormatBool(smartDnsProxyingEnabled),
				},
			},
			TrustDomain: trustDomain,
		}
		yaml, err := protomarshal.ToYAML(meshConfig)
		Expect(err).To(BeNil())
		return corev1sets.NewConfigMapSet(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   meshNs,
				Name:        "istio",
				ClusterName: clusterName,
			},
			Data: map[string]string{
				"mesh": yaml,
			},
		})
	}

	settings := &settingsv1alpha2.DiscoverySettings{}

	It("does not detect Istio when it is not there", func() {

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: "a", Name: "a"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "test-image",
							},
						},
					},
				},
			},
		}

		detector := NewMeshDetector(
			ctx,
		)

		inRemote := input.NewInputDiscoveryInputSnapshotManualBuilder("")
		inRemote.AddDeployments([]*appsv1.Deployment{deployment})

		meshes, err := detector.DetectMeshes(inRemote.Build(), settings)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshes).To(HaveLen(0))
	})

	It("detects a mesh from a deployment named istio-pilot", func() {
		configMaps := istioConfigMap()
		deployment := istioDeployment(pilotDeploymentName)

		detector := NewMeshDetector(
			ctx,
		)

		inRemote := input.NewInputDiscoveryInputSnapshotManualBuilder("")
		inRemote.AddDeployments([]*appsv1.Deployment{deployment})
		inRemote.AddConfigMaps(configMaps.List())

		meshes, err := detector.DetectMeshes(inRemote.Build(), settings)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshes).To(HaveLen(1))
		Expect(meshes[0]).To(Equal(&discoveryv1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-pilot-namespace-cluster",
				Namespace: defaults.GetPodNamespace(),
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{Istio: &discoveryv1alpha2.MeshSpec_Istio{
					SmartDnsProxyingEnabled: smartDnsProxyingEnabled,
					Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
						PodLabels: map[string]string{"app": "istiod"},
					},
					CitadelInfo: &discoveryv1alpha2.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
				}},
			},
		}))
	})

	It("detects a mesh from a deployment named istiod", func() {
		configMaps := istioConfigMap()
		deployment := istioDeployment(istiodDeploymentName)

		inRemote := input.NewInputDiscoveryInputSnapshotManualBuilder("")
		inRemote.AddDeployments([]*appsv1.Deployment{deployment})
		inRemote.AddConfigMaps(configMaps.List())

		detector := NewMeshDetector(
			ctx,
		)

		meshes, err := detector.DetectMeshes(inRemote.Build(), settings)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshes).To(HaveLen(1))
		Expect(meshes[0]).To(Equal(&discoveryv1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istiod-namespace-cluster",
				Namespace: defaults.GetPodNamespace(),
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{Istio: &discoveryv1alpha2.MeshSpec_Istio{
					SmartDnsProxyingEnabled: smartDnsProxyingEnabled,
					Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						PodLabels: map[string]string{"app": "istiod"},
						Version:   "latest",
					},
					CitadelInfo: &discoveryv1alpha2.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
				}},
			},
		}))
	})

	It("detects a ingress gateway which uses a nodeport service", func() {
		configMaps := istioConfigMap()

		istioNamespace := defaults.GetPodNamespace()

		workloadLabels := map[string]string{"istio": "ingressgateway"}
		services := corev1sets.NewServiceSet(&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ingress-svc",
				Namespace:   meshNs,
				ClusterName: clusterName,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:     "tls",
						Protocol: "TCP",
						Port:     1234,
						NodePort: 5678,
					},
					{
						Name:     "https",
						Protocol: "HTTPS",
						Port:     2345,
						NodePort: 6789,
					}},
				Selector: workloadLabels,
				Type:     corev1.ServiceTypeNodePort,
			},
		})

		nodeName := "ingress-node"
		pods := corev1sets.NewPodSet(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ingress-pod",
				Namespace:   meshNs,
				ClusterName: clusterName,
				Labels:      workloadLabels,
			},
			Spec: corev1.PodSpec{
				NodeName: nodeName,
			},
		})
		nodes := corev1sets.NewNodeSet(&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:        nodeName,
				ClusterName: clusterName,
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{
						Type:    corev1.NodeInternalDNS,
						Address: "internal.domain",
					},
					{
						Type:    corev1.NodeExternalDNS,
						Address: "external.domain",
					},
				},
			},
		})

		detector := NewMeshDetector(
			ctx,
		)

		deployment := istioDeployment(istiodDeploymentName)

		inRemote := input.NewInputDiscoveryInputSnapshotManualBuilder("")
		inRemote.AddDeployments([]*appsv1.Deployment{deployment})
		inRemote.AddConfigMaps(configMaps.List())
		inRemote.AddServices(services.List())
		inRemote.AddPods(pods.List())
		inRemote.AddNodes(nodes.List())

		meshes, err := detector.DetectMeshes(inRemote.Build(), settings)
		Expect(err).NotTo(HaveOccurred())

		expectedMesh := &discoveryv1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istiod-namespace-cluster",
				Namespace: istioNamespace,
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{Istio: &discoveryv1alpha2.MeshSpec_Istio{
					SmartDnsProxyingEnabled: smartDnsProxyingEnabled,
					Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
						PodLabels: map[string]string{"app": "istiod"},
					},
					CitadelInfo: &discoveryv1alpha2.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
					IngressGateways: []*discoveryv1alpha2.MeshSpec_Istio_IngressGatewayInfo{{
						WorkloadLabels:   workloadLabels,
						ExternalAddress:  "external.domain",
						ExternalTlsPort:  5678,
						TlsContainerPort: 1234,
					}},
				}},
			},
		}

		Expect(meshes).To(HaveLen(1))
		Expect(meshes[0]).To(Equal(expectedMesh))
	})

	It("uses settings to detect ingress gateways", func() {
		configMaps := istioConfigMap()
		workloadLabels := map[string]string{"mykey": "myvalue"}
		services := corev1sets.NewServiceSet(&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ingress-svc",
				Namespace:   meshNs,
				ClusterName: clusterName,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:     "specialport",
					Protocol: "TCP",
					Port:     1234,
					NodePort: 5678,
				}},
				Selector: workloadLabels,
				Type:     corev1.ServiceTypeNodePort,
			},
		})

		nodeName := "ingress-node"
		pods := corev1sets.NewPodSet(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ingress-pod",
				Namespace:   meshNs,
				ClusterName: clusterName,
				Labels:      workloadLabels,
			},
			Spec: corev1.PodSpec{
				NodeName: nodeName,
			},
		})
		nodes := corev1sets.NewNodeSet(&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:        nodeName,
				ClusterName: clusterName,
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{
						Type:    corev1.NodeInternalDNS,
						Address: "internal.domain",
					},
					{
						Type:    corev1.NodeExternalDNS,
						Address: "external.domain",
					},
				},
			},
		})

		detector := NewMeshDetector(
			ctx,
		)

		deployment := istioDeployment(istiodDeploymentName)

		inRemote := input.NewInputDiscoveryInputSnapshotManualBuilder("")
		inRemote.AddDeployments([]*appsv1.Deployment{deployment})
		inRemote.AddConfigMaps(configMaps.List())
		inRemote.AddServices(services.List())
		inRemote.AddPods(pods.List())
		inRemote.AddNodes(nodes.List())
		settings := &settingsv1alpha2.DiscoverySettings{
			Istio: &settingsv1alpha2.DiscoverySettings_Istio{
				IngressGatewayDetectors: map[string]*settingsv1alpha2.DiscoverySettings_Istio_IngressGatewayDetector{
					"*": {
						GatewayWorkloadLabels: map[string]string{"mykey": "myvalue"},
						GatewayTlsPortName:    "myport",
					},
					clusterName: {
						GatewayTlsPortName: "specialport",
					},
				},
			},
		}

		meshes, err := detector.DetectMeshes(inRemote.Build(), settings)
		Expect(err).NotTo(HaveOccurred())

		expectedMesh := &discoveryv1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istiod-namespace-cluster",
				Namespace: defaults.GetPodNamespace(),
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{Istio: &discoveryv1alpha2.MeshSpec_Istio{
					SmartDnsProxyingEnabled: smartDnsProxyingEnabled,
					Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
						PodLabels: map[string]string{"app": "istiod"},
					},
					CitadelInfo: &discoveryv1alpha2.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
					IngressGateways: []*discoveryv1alpha2.MeshSpec_Istio_IngressGatewayInfo{{
						WorkloadLabels:    workloadLabels,
						ExternalAddress:   "external.domain",
						ExternalTlsPort:   5678,
						TlsContainerPort:  1234,
						ExternalHttpsPort: 6789,
						HttpsPort:         2345,
					}},
				}},
			},
		}

		Expect(meshes).To(HaveLen(1))
		Expect(meshes[0]).To(Equal(expectedMesh))
	})

	It("detects a egress gateway", func() {
		configMaps := istioConfigMap()

		istioNamespace := defaults.GetPodNamespace()

		ingressLabels := map[string]string{"istio": "ingressgateway"}
		ingressService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ingress-svc",
				Namespace:   meshNs,
				ClusterName: clusterName,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:     "tls",
						Protocol: "TCP",
						Port:     1234,
						NodePort: 5678,
					},
					{
						Name:     "https",
						Protocol: "HTTPS",
						Port:     2345,
						NodePort: 6789,
					}},
				Selector: ingressLabels,
				Type:     corev1.ServiceTypeNodePort,
			},
		}

		egressLabels := map[string]string{"istio": "egressgateway"}
		egressService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "egress-svc",
				Namespace:   meshNs,
				ClusterName: clusterName,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:     "tls",
						Protocol: "TCP",
						Port:     1234,
					},
					{
						Name:     "https",
						Protocol: "HTTPS",
						Port:     2345,
					}},
				Selector: egressLabels,
				Type:     corev1.ServiceTypeClusterIP,
			},
		}
		services := corev1sets.NewServiceSet(ingressService, egressService)

		ingressNodeName := "ingress-node"
		egressNodeName := "egress-node"
		pods := corev1sets.NewPodSet(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "ingress-pod",
					Namespace:   meshNs,
					ClusterName: clusterName,
					Labels:      ingressLabels,
				},
				Spec: corev1.PodSpec{
					NodeName: ingressNodeName,
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "egress-pod",
					Namespace:   meshNs,
					ClusterName: clusterName,
					Labels:      egressLabels,
				},
				Spec: corev1.PodSpec{
					NodeName: egressNodeName,
				},
			},
		)
		nodes := corev1sets.NewNodeSet(
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        ingressNodeName,
					ClusterName: clusterName,
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalDNS,
							Address: "internal.domain",
						},
						{
							Type:    corev1.NodeExternalDNS,
							Address: "external.domain",
						},
					},
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        egressNodeName,
					ClusterName: clusterName,
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalDNS,
							Address: "internal.domain",
						},
						{
							Type:    corev1.NodeExternalDNS,
							Address: "external.domain",
						},
					},
				},
			},
		)

		detector := NewMeshDetector(
			ctx,
		)

		deployment := istioDeployment(istiodDeploymentName)

		in := input.NewInputRemoteSnapshotManualBuilder("")
		in.AddDeployments([]*appsv1.Deployment{deployment})
		in.AddConfigMaps(configMaps.List())
		in.AddServices(services.List())
		in.AddPods(pods.List())
		in.AddNodes(nodes.List())

		meshes, err := detector.DetectMeshes(in.Build())
		Expect(err).NotTo(HaveOccurred())

		expectedMesh := &v1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istiod-namespace-cluster",
				Namespace: istioNamespace,
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: v1alpha2.MeshSpec{
				MeshType: &v1alpha2.MeshSpec_Istio_{Istio: &v1alpha2.MeshSpec_Istio{
					Installation: &v1alpha2.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
						PodLabels: map[string]string{"app": "istiod"},
					},
					CitadelInfo: &v1alpha2.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
					IngressGateways: []*v1alpha2.MeshSpec_Istio_IngressGatewayInfo{{
						WorkloadLabels:    ingressLabels,
						ExternalAddress:   "external.domain",
						ExternalTlsPort:   5678,
						TlsContainerPort:  1234,
						ExternalHttpsPort: 6789,
						HttpsPort:         2345,
					}},
					EgressGateways: []*v1alpha2.MeshSpec_Istio_EgressGatewayInfo{{
						Name:           "egress-svc",
						WorkloadLabels: egressLabels,
						TlsPort:        1234,
						HttpsPort:      2345,
					}},
				}},
			},
		}

		Expect(meshes).To(HaveLen(1))
		Expect(meshes[0]).To(Equal(expectedMesh))
	})

})
