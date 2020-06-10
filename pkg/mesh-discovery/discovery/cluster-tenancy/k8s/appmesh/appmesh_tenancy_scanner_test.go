package appmesh_tenancy_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser"
	mock_aws "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser/mocks"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	appmesh_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s/appmesh"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
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
		mockRemoteClient        *mock_controller_runtime.MockClient
		mockAwsAccountIdFetcher *mock_aws.MockAwsAccountIdFetcher
		appMeshTenancyRegistrar k8s_tenancy.ClusterTenancyRegistrar
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockAppMeshParser = mock_aws.NewMockAppMeshScanner(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockRemoteClient = mock_controller_runtime.NewMockClient(ctrl)
		mockAwsAccountIdFetcher = mock_aws.NewMockAwsAccountIdFetcher(ctrl)
		appMeshTenancyRegistrar = appmesh_tenancy.NewAppmeshTenancyScanner(
			mockAppMeshParser,
			mockMeshClient,
			mockRemoteClient,
			mockAwsAccountIdFetcher,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return Mesh for pod", func() {
		expectedMesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
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
		mockAwsAccountIdFetcher.EXPECT().GetEksAccountId(ctx, mockRemoteClient).Return(aws_utils.AwsAccountId(appMeshPod.AwsAccountID), nil)
		mockAppMeshParser.EXPECT().ScanPodForAppMesh(pod, aws_utils.AwsAccountId(appMeshPod.AwsAccountID)).Return(appMeshPod, nil)
		mockMeshClient.EXPECT().GetMesh(ctx, client.ObjectKey{
			Name:      metadata.BuildAppMeshName(appMeshPod.AppMeshName, appMeshPod.Region, appMeshPod.AwsAccountID),
			Namespace: container_runtime.GetWriteNamespace(),
		}).Return(expectedMesh, nil)
		mesh, err := appMeshTenancyRegistrar.MeshFromSidecar(ctx, pod)
		Expect(err).ToNot(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("should register cluster for Mesh", func() {
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
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
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b", clusterName},
					},
				},
			},
		}
		err := appMeshTenancyRegistrar.RegisterMesh(ctx, clusterName, mesh)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deregister cluster for Mesh", func() {
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
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
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b"},
					},
				},
			},
		}
		err := appMeshTenancyRegistrar.DeregisterMesh(ctx, clusterName, mesh)
		Expect(err).ToNot(HaveOccurred())
	})
})
