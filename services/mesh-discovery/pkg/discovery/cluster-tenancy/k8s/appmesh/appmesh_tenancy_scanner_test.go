package appmesh_tenancy_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	v1 "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser/mocks"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	appmesh_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s/appmesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/appmesh"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	"github.com/solo-io/skv2/pkg/utils"
	k8s_core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AppmeshTenancyFinder", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     = context.TODO()
		clusterName             = "test-cluster-name"
		mockAppMeshParser       *mock_aws.MockAppMeshScanner
		mockMeshClient          *mock_core.MockMeshClient
		mockConfigMapClient     *mock_kubernetes_core.MockConfigMapClient
		mockRemoteClient        *mock_controller_runtime.MockClient
		appMeshTenancyRegistrar k8s_tenancy.ClusterTenancyRegistrar
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockAppMeshParser = mock_aws.NewMockAppMeshScanner(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockConfigMapClient = mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		mockRemoteClient = mock_controller_runtime.NewMockClient(ctrl)
		appMeshTenancyRegistrar = appmesh_tenancy.NewAppmeshTenancyScanner(
			mockAppMeshParser,
			func(client client.Client) v1.ConfigMapClient {
				return mockConfigMapClient
			},
			mockMeshClient,
			mockRemoteClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return Mesh for pod", func() {
		expectedMesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b", clusterName},
					},
				},
			},
		}
		pod := &k8s_core.Pod{}
		appMeshPod := &aws_utils.AppMeshPod{
			AwsAccountID:    "aws-account-id",
			Region:          "region",
			AppMeshName:     "appmeshname",
			VirtualNodeName: "virtualnodename",
		}
		configMap := &k8s_core.ConfigMap{}
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, appmesh.AwsAuthConfigMapKey).Return(configMap, nil)
		mockAppMeshParser.EXPECT().ScanPodForAppMesh(pod, configMap).Return(appMeshPod, nil)
		mockMeshClient.EXPECT().GetMesh(ctx, client.ObjectKey{
			Name:      metadata.BuildAppMeshName(appMeshPod.AppMeshName, appMeshPod.Region, appMeshPod.AwsAccountID),
			Namespace: env.GetWriteNamespace(),
		}).Return(expectedMesh, nil)
		mesh, err := appMeshTenancyRegistrar.MeshFromSidecar(ctx, pod)
		Expect(err).ToNot(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("should register cluster for Mesh", func() {
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b"},
					},
				},
			},
		}
		// Update
		updatedMesh := mesh.DeepCopy()
		updatedMesh.Spec.GetAwsAppMesh().Clusters = append(updatedMesh.Spec.GetAwsAppMesh().Clusters, clusterName)
		mockMeshClient.EXPECT().UpdateMesh(ctx, updatedMesh).Return(nil)
		err := appMeshTenancyRegistrar.RegisterMesh(ctx, clusterName, mesh)
		Expect(err).ToNot(HaveOccurred())
	})

	It("while registering cluster, it should update Mesh only if needed", func() {
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b", clusterName},
					},
				},
			},
		}
		err := appMeshTenancyRegistrar.RegisterMesh(ctx, clusterName, mesh)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deregister cluster for Mesh", func() {
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b", clusterName},
					},
				},
			},
		}
		// Update
		updatedMesh := mesh.DeepCopy()
		updatedMesh.Spec.GetAwsAppMesh().Clusters = utils.RemoveString(updatedMesh.Spec.GetAwsAppMesh().Clusters, clusterName)
		mockMeshClient.EXPECT().UpdateMesh(ctx, updatedMesh).Return(nil)
		err := appMeshTenancyRegistrar.DeregisterMesh(ctx, clusterName, mesh)
		Expect(err).ToNot(HaveOccurred())
	})

	It("while deregistering cluster, it should update Mesh only if needed", func() {
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b"},
					},
				},
			},
		}
		err := appMeshTenancyRegistrar.DeregisterMesh(ctx, clusterName, mesh)
		Expect(err).ToNot(HaveOccurred())
	})
})
