package istio_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	mock_docker "github.com/solo-io/service-mesh-hub/pkg/common/docker/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/istio"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Istio Mesh Scanner", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 = context.TODO()
		istioNs             = "istio-system"
		clusterScopedClient client.Client
		mockImageNameParser *mock_docker.MockImageNameParser
		mockConfigMapClient *mock_kubernetes_core.MockConfigMapClient
		istioMeshScanner    istio.IstioMeshScanner
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockImageNameParser = mock_docker.NewMockImageNameParser(ctrl)
		mockConfigMapClient = mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		istioMeshScanner = istio.NewIstioMeshScanner(
			mockImageNameParser,
			func(client client.Client) kubernetes_core.ConfigMapClient {
				return mockConfigMapClient
			})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not detect Istio when it is not there", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "test-image",
							},
						},
					},
				},
			},
		}
		mesh, err := istioMeshScanner.ScanDeployment(ctx, deployment, clusterScopedClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("reports an error when the image name is unparseable", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, ClusterName: "test-cluster", Name: istio.IstiodDeploymentName},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "istio-pilot:latest",
							},
						},
					},
				},
			},
		}
		testErr := eris.New("test-err")
		mockImageNameParser.
			EXPECT().
			Parse("istio-pilot:latest").
			Return(nil, testErr)
		mesh, err := istioMeshScanner.ScanDeployment(ctx, deployment, clusterScopedClient)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(mesh).To(BeNil())
	})

	It("discovers Istiod deployment", func() {
		serviceAccountName := "service-account-name"
		trustDomain := "cluster.local"
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, ClusterName: "test-cluster", Name: istio.IstiodDeploymentName},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "istio-pilot:latest",
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
			},
		}
		expectedMesh := &discoveryv1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-istio-system-test-cluster",
				Namespace: env.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels,
			},
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						Installation: &discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               "latest",
						},
						CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{
							TrustDomain:           trustDomain,
							CitadelNamespace:      istioNs,
							CitadelServiceAccount: serviceAccountName,
						},
					},
				},
				Cluster: &core_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: env.GetWriteNamespace(),
				},
			},
		}
		mockImageNameParser.
			EXPECT().
			Parse("istio-pilot:latest").
			Return(&docker.Image{
				Domain: "test.com",
				Path:   "istio",
				Tag:    "latest",
			}, nil)
		configMap := &kubev1.ConfigMap{
			Data: map[string]string{
				"mesh": fmt.Sprintf("trustDomain: \"%s\"", trustDomain),
			},
		}
		mockConfigMapClient.
			EXPECT().
			GetConfigMap(ctx, client.ObjectKey{Name: istio.IstioConfigMapName, Namespace: istioNs}).
			Return(configMap, nil)
		mesh, err := istioMeshScanner.ScanDeployment(ctx, deployment, clusterScopedClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("discovers istio-citadel deployment", func() {
		serviceAccountName := "service-account-name"
		trustDomain := "cluster.local"
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, ClusterName: "test-cluster", Name: istio.CitadelDeploymentName},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "istio-citadel:latest",
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
			},
		}
		expectedMesh := &discoveryv1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-istio-system-test-cluster",
				Namespace: env.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels,
			},
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						Installation: &discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               "latest",
						},
						CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{
							TrustDomain:           trustDomain,
							CitadelNamespace:      istioNs,
							CitadelServiceAccount: serviceAccountName,
						},
					},
				},
				Cluster: &core_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: env.GetWriteNamespace(),
				},
			},
		}
		mockImageNameParser.
			EXPECT().
			Parse("istio-citadel:latest").
			Return(&docker.Image{
				Domain: "test.com",
				Path:   "istio",
				Tag:    "latest",
			}, nil)
		configMap := &kubev1.ConfigMap{
			Data: map[string]string{
				"mesh": fmt.Sprintf("trustDomain: \"%s\"", trustDomain),
			},
		}
		mockConfigMapClient.
			EXPECT().
			GetConfigMap(ctx, client.ObjectKey{Name: istio.IstioConfigMapName, Namespace: istioNs}).
			Return(configMap, nil)
		mesh, err := istioMeshScanner.ScanDeployment(ctx, deployment, clusterScopedClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})
})
