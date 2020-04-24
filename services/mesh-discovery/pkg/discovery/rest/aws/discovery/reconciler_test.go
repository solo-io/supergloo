package discovery_test

import (
	"context"
	"fmt"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws/discovery"
	mock_appmesh_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/aws/appmesh"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Reconciler", func() {
	var (
		ctrl                       *gomock.Controller
		ctx                        context.Context
		mockMeshClient             *mock_core.MockMeshClient
		mockAppMeshClient          *mock_appmesh_clients.MockAppMeshAPI
		meshPlatformName           string
		appMeshDiscoveryReconciler rest.RestAPIDiscoveryReconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		meshPlatformName = "aws-account-name"
		mockAppMeshClient = mock_appmesh_clients.NewMockAppMeshAPI(ctrl)
		appMeshDiscoveryReconciler = discovery.NewAppMeshDiscoveryReconciler(
			mockMeshClient,
			mockAppMeshClient,
			meshPlatformName,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectReconcileMeshes = func() {
		page1Input := &appmesh.ListMeshesInput{
			Limit: discovery.NumItemsPerRequest,
		}
		page2Input := &appmesh.ListMeshesInput{
			Limit:     discovery.NumItemsPerRequest,
			NextToken: aws2.String("page-2-token"),
		}
		meshRefs := []*appmesh.MeshRef{
			{
				MeshName: aws2.String("mesh-name-1"),
			},
			{
				MeshName: aws2.String("mesh-name-2"),
			},
			{
				MeshName: aws2.String("mesh-name-3"),
			},
			{
				MeshName: aws2.String("mesh-name-4"),
			},
			{
				MeshName: aws2.String("mesh-name-5"),
			},
			{
				MeshName: aws2.String("mesh-name-6"),
			},
		}
		page1 := &appmesh.ListMeshesOutput{
			Meshes:    meshRefs[:3],
			NextToken: page2Input.NextToken,
		}
		page2 := &appmesh.ListMeshesOutput{
			Meshes:    meshRefs[3:],
			NextToken: nil,
		}
		mockAppMeshClient.EXPECT().ListMeshes(page1Input).Return(page1, nil)
		mockAppMeshClient.EXPECT().ListMeshes(page2Input).Return(page2, nil)
		for _, meshRef := range meshRefs {
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s", discovery.ObjectNamePrefix, *meshRef.MeshName, meshPlatformName),
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
						AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
							Name:           *meshRef.MeshName,
							AwsAccountName: meshPlatformName,
							Region:         aws.Region,
						},
					},
				},
			}
			mockMeshClient.EXPECT().UpsertMeshSpec(ctx, mesh).Return(nil)
		}
		existingMeshes := &zephyr_discovery.MeshList{
			Items: []zephyr_discovery.Mesh{
				{ObjectMeta: k8s_meta_types.ObjectMeta{ // should not be deleted
					Name: fmt.Sprintf("%s-%s-%s", discovery.ObjectNamePrefix, *meshRefs[0].MeshName, meshPlatformName)},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "non-existent-1"},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}}}, // should be deleted
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "non-existent-2"},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}}}, // should be deleted
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(existingMeshes, nil)
		for _, existingMesh := range existingMeshes.Items[1:] {
			existingMesh := existingMesh
			mockMeshClient.EXPECT().DeleteMesh(ctx, clients.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
	}

	It("should reconcile Meshes", func() {
		expectReconcileMeshes()
		err := appMeshDiscoveryReconciler.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
	})
})
