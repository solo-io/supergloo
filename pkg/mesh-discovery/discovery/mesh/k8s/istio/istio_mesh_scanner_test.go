package istio_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker"
	mock_docker "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/istio"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Istio Mesh Scanner", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 = context.TODO()
		istioNs             = "istio-system"
		clusterName         = "test-cluster-name"
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
			func(client client.Client) k8s_core.ConfigMapClient {
				return mockConfigMapClient
			})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not detect Istio when it is not there", func() {
		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: istioNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Image: "test-image",
							},
						},
					},
				},
			},
		}
		mesh, err := istioMeshScanner.ScanDeployment(ctx, clusterName, deployment, clusterScopedClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("reports an error when the image name is unparseable", func() {
		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: istioNs, Name: istio.IstiodDeploymentName},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
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
		mesh, err := istioMeshScanner.ScanDeployment(ctx, clusterName, deployment, clusterScopedClient)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(mesh).To(BeNil())
	})

	It("discovers Istiod deployment", func() {
		serviceAccountName := "service-account-name"
		trustDomain := "cluster.local"
		imageVersion := "1.5.minor-version"
		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: istioNs, ClusterName: clusterName, Name: istio.IstiodDeploymentName},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Image: "istio-pilot:" + imageVersion,
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
			},
		}
		expectedMesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "istio-istio-system-" + clusterName,
				Namespace: container_runtime.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels,
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: deployment.GetNamespace(),
								Version:               imageVersion,
							},
							CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
								TrustDomain:           trustDomain,
								CitadelNamespace:      istioNs,
								CitadelServiceAccount: serviceAccountName,
							},
						},
					},
				},
				Cluster: &smh_core_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: container_runtime.GetWriteNamespace(),
				},
			},
		}
		mockImageNameParser.
			EXPECT().
			Parse("istio-pilot:"+imageVersion).
			Return(&docker.Image{
				Domain: "test.com",
				Path:   "istio",
				Tag:    imageVersion,
			}, nil)
		configMap := &k8s_core_types.ConfigMap{
			Data: map[string]string{
				"mesh": fmt.Sprintf("trustDomain: \"%s\"", trustDomain),
			},
		}
		mockConfigMapClient.
			EXPECT().
			GetConfigMap(ctx, client.ObjectKey{Name: istio.IstioConfigMapName, Namespace: istioNs}).
			Return(configMap, nil)
		mesh, err := istioMeshScanner.ScanDeployment(ctx, clusterName, deployment, clusterScopedClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("discovers istio-citadel deployment", func() {
		serviceAccountName := "service-account-name"
		trustDomain := "cluster.local"
		imageVersion := "1.6.minor-version"
		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: istioNs, ClusterName: clusterName, Name: istio.CitadelDeploymentName},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Image: "istio-citadel:" + imageVersion,
							},
						},
						ServiceAccountName: serviceAccountName,
					},
				},
			},
		}
		expectedMesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "istio-istio-system-" + clusterName,
				Namespace: container_runtime.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels,
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_6_{
					Istio1_6: &smh_discovery_types.MeshSpec_Istio1_6{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: deployment.GetNamespace(),
								Version:               imageVersion,
							},
							CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
								TrustDomain:           trustDomain,
								CitadelNamespace:      istioNs,
								CitadelServiceAccount: serviceAccountName,
							},
						},
					},
				},
				Cluster: &smh_core_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: container_runtime.GetWriteNamespace(),
				},
			},
		}
		mockImageNameParser.
			EXPECT().
			Parse("istio-citadel:"+imageVersion).
			Return(&docker.Image{
				Domain: "test.com",
				Path:   "istio",
				Tag:    imageVersion,
			}, nil)
		configMap := &k8s_core_types.ConfigMap{
			Data: map[string]string{
				"mesh": fmt.Sprintf("trustDomain: \"%s\"", trustDomain),
			},
		}
		mockConfigMapClient.
			EXPECT().
			GetConfigMap(ctx, client.ObjectKey{Name: istio.IstioConfigMapName, Namespace: istioNs}).
			Return(configMap, nil)
		mesh, err := istioMeshScanner.ScanDeployment(ctx, clusterName, deployment, clusterScopedClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})
})
