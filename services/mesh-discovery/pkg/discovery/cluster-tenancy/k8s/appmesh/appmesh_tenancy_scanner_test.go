package appmesh_tenancy_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	appmesh_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s/appmesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	k8s_core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AppmeshTenancyFinder", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   = context.TODO()
		clusterName           = "test-cluster-name"
		mockAppMeshParser     *mock_aws.MockAppMeshParser
		mockMeshClient        *mock_core.MockMeshClient
		appMeshTenancyScanner k8s_tenancy.ClusterTenancyScanner
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockAppMeshParser = mock_aws.NewMockAppMeshParser(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		appMeshTenancyScanner = appmesh_tenancy.NewAppmeshTenancyScanner(mockAppMeshParser, mockMeshClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should update mesh tenancy", func() {
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
						Clusters: []string{"a", "b"},
					},
				},
			},
		}
		pod := &k8s_core.Pod{}
		appMeshPod := &aws.AppMeshPod{
			AwsAccountID:    "aws-account-id",
			Region:          "region",
			AppMeshName:     "appmeshname",
			VirtualNodeName: "virtualnodename",
		}
		mockAppMeshParser.EXPECT().ScanPodForAppMesh(pod).Return(appMeshPod, nil)
		mockMeshClient.EXPECT().GetMesh(ctx, client.ObjectKey{
			Name:      metadata.BuildAppMeshName(appMeshPod.AppMeshName, appMeshPod.Region, appMeshPod.AwsAccountID),
			Namespace: env.GetWriteNamespace(),
		}).Return(mesh, nil)
		// Update
		updatedMesh := mesh.DeepCopy()
		updatedMesh.Spec.GetAwsAppMesh().Clusters = append(updatedMesh.Spec.GetAwsAppMesh().Clusters, clusterName)
		mockMeshClient.EXPECT().UpdateMesh(ctx, updatedMesh).Return(nil)
		appMeshTenancyScanner.UpdateMeshTenancy(ctx, clusterName, pod)
	})

	It("should not update mesh tenancy if not found", func() {
		pod := &k8s_core.Pod{}
		mockAppMeshParser.EXPECT().ScanPodForAppMesh(pod).Return(nil, nil)
		err := appMeshTenancyScanner.UpdateMeshTenancy(ctx, clusterName, pod)
		Expect(err).To(BeNil())
	})
})
