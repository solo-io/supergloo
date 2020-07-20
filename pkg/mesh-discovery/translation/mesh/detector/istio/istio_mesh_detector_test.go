package istio_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/common/defaults"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/detector/istio"
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
			},
		}
	}

	trustDomain := "cluster.local"
	istioConfigMap := func() corev1sets.ConfigMapSet {
		return corev1sets.NewConfigMapSet(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   meshNs,
				Name:        "istio",
				ClusterName: clusterName,
			},
			Data: map[string]string{
				"mesh": fmt.Sprintf("trustDomain: \"%s\"", trustDomain),
			},
		})
	}

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

		configMaps := corev1sets.NewConfigMapSet()
		services := corev1sets.NewServiceSet()
		pods := corev1sets.NewPodSet()
		nodes := corev1sets.NewNodeSet()

		detector := NewMeshDetector(
			ctx,
			configMaps,
			services,
			pods,
			nodes,
		)

		mesh, err := detector.DetectMesh(deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("detects a mesh from a deployment named istio-pilot", func() {
		configMaps := istioConfigMap()

		services := corev1sets.NewServiceSet()
		pods := corev1sets.NewPodSet()
		nodes := corev1sets.NewNodeSet()

		detector := NewMeshDetector(
			ctx,
			configMaps,
			services,
			pods,
			nodes,
		)

		deployment := istioDeployment(pilotDeploymentName)
		mesh, err := detector.DetectMesh(deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(&v1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-pilot-namespace-cluster",
				Namespace: defaults.GetPodNamespace(),
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: v1alpha1.MeshSpec{
				MeshType: &v1alpha1.MeshSpec_Istio_{Istio: &v1alpha1.MeshSpec_Istio{
					Installation: &v1alpha1.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
					},
					CitadelInfo: &v1alpha1.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
				}},
			},
		}))
	})

	It("detects a mesh from a deployment named istiod", func() {
		configMaps := istioConfigMap()

		services := corev1sets.NewServiceSet()
		pods := corev1sets.NewPodSet()
		nodes := corev1sets.NewNodeSet()

		detector := NewMeshDetector(
			ctx,
			configMaps,
			services,
			pods,
			nodes,
		)

		deployment := istioDeployment(istiodDeploymentName)
		mesh, err := detector.DetectMesh(deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(&v1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istiod-namespace-cluster",
				Namespace: defaults.GetPodNamespace(),
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: v1alpha1.MeshSpec{
				MeshType: &v1alpha1.MeshSpec_Istio_{Istio: &v1alpha1.MeshSpec_Istio{
					Installation: &v1alpha1.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
					},
					CitadelInfo: &v1alpha1.MeshSpec_Istio_CitadelInfo{
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
				Ports: []corev1.ServicePort{{
					Name:     "tls",
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
			configMaps,
			services,
			pods,
			nodes,
		)

		deployment := istioDeployment(istiodDeploymentName)
		mesh, err := detector.DetectMesh(deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(&v1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istiod-namespace-cluster",
				Namespace: istioNamespace,
				Labels:    labelutils.ClusterLabels(clusterName),
			},
			Spec: v1alpha1.MeshSpec{
				MeshType: &v1alpha1.MeshSpec_Istio_{Istio: &v1alpha1.MeshSpec_Istio{
					Installation: &v1alpha1.MeshSpec_MeshInstallation{
						Namespace: meshNs,
						Cluster:   clusterName,
						Version:   "latest",
					},
					CitadelInfo: &v1alpha1.MeshSpec_Istio_CitadelInfo{
						TrustDomain:           trustDomain,
						CitadelServiceAccount: serviceAccountName,
					},
					IngressGateways: []*v1alpha1.MeshSpec_Istio_IngressGatewayInfo{{
						WorkloadLabels:   workloadLabels,
						ExternalAddress:  "external.domain",
						ExternalTlsPort:  5678,
						TlsContainerPort: 1234,
					}},
				}},
			},
		}))
	})

})
