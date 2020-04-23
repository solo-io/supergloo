package aws_test

import (
	"context"
	"fmt"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws"
	mock_appmesh_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/aws/appmesh"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Reconciler", func() {
	var (
		ctrl                       *gomock.Controller
		ctx                        context.Context
		mockMeshClient             *mock_core.MockMeshClient
		mockMeshWorkloadClient     *mock_core.MockMeshWorkloadClient
		mockMeshServiceClient      *mock_core.MockMeshServiceClient
		mockAppMeshClient          *mock_appmesh_clients.MockAppMeshAPI
		meshPlatformName           string
		appMeshDiscoveryReconciler rest_api.RestAPIDiscoveryReconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		meshPlatformName = "aws-account-name"
		mockAppMeshClient = mock_appmesh_clients.NewMockAppMeshAPI(ctrl)
		appMeshDiscoveryReconciler = aws.NewAppMeshDiscoveryReconciler(
			mockMeshClient,
			mockMeshWorkloadClient,
			mockMeshServiceClient,
			mockAppMeshClient,
			meshPlatformName,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectReconcileMeshes = func() map[string]*zephyr_discovery.Mesh {
		discoveredMeshes := make(map[string]*zephyr_discovery.Mesh)
		page1Input := &appmesh.ListMeshesInput{
			Limit: aws.NumItemsPerRequest,
		}
		page2Input := &appmesh.ListMeshesInput{
			Limit:     aws.NumItemsPerRequest,
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
					Name:      fmt.Sprintf("%s-%s-%s", aws.ObjectNamePrefix, *meshRef.MeshName, meshPlatformName),
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
			discoveredMeshes[*meshRef.MeshName] = mesh
			mockMeshClient.EXPECT().UpsertMeshSpec(ctx, mesh).Return(nil)
		}
		existingMeshes := &zephyr_discovery.MeshList{
			Items: []zephyr_discovery.Mesh{
				{ObjectMeta: k8s_meta_types.ObjectMeta{ // should not be deleted
					Name: fmt.Sprintf("%s-%s-%s", aws.ObjectNamePrefix, *meshRefs[0].MeshName, meshPlatformName)},
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
		return discoveredMeshes
	}

	var expectReconcileMeshServices = func(meshes map[string]*zephyr_discovery.Mesh) {
		for appMeshName, parentMesh := range meshes {
			page1Input := &appmesh.ListVirtualServicesInput{
				Limit:    aws.NumItemsPerRequest,
				MeshName: aws2.String(appMeshName),
			}
			page2Input := &appmesh.ListVirtualServicesInput{
				Limit:     aws.NumItemsPerRequest,
				MeshName:  aws2.String(appMeshName),
				NextToken: aws2.String("page-2-token"),
			}
			virtualServiceRefs := []*appmesh.VirtualServiceRef{
				{
					MeshName: aws2.String(appMeshName),
				},
				{
					MeshName: aws2.String(appMeshName),
				},
				{
					MeshName: aws2.String(appMeshName),
				},
				{
					MeshName: aws2.String(appMeshName),
				},
				{
					MeshName: aws2.String(appMeshName),
				},
				{
					MeshName: aws2.String(appMeshName),
				},
			}
			page1 := &appmesh.ListVirtualServicesOutput{
				VirtualServices: virtualServiceRefs[:3],
				NextToken:       page2Input.NextToken,
			}
			page2 := &appmesh.ListVirtualServicesOutput{
				VirtualServices: virtualServiceRefs[3:],
				NextToken:       nil,
			}
			mockAppMeshClient.EXPECT().ListVirtualServices(page1Input).Return(page1, nil)
			mockAppMeshClient.EXPECT().ListVirtualServices(page2Input).Return(page2, nil)
			for _, virtualServiceRef := range virtualServiceRefs {
				meshService := &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: fmt.Sprintf("%s-%s-%s-%s",
							aws.ObjectNamePrefix,
							aws2.StringValue(virtualServiceRef.VirtualServiceName),
							aws2.StringValue(virtualServiceRef.MeshName),
							meshPlatformName),
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: &zephyr_core_types.ResourceRef{
							Name:      parentMesh.GetName(),
							Namespace: parentMesh.GetNamespace(),
							Cluster:   parentMesh.Spec.GetAwsAppMesh().GetAwsAccountName(),
						},
					},
				}
				mockMeshServiceClient.EXPECT().UpsertMeshServiceSpec(ctx, meshService).Return(nil)
			}
			existingMeshServices := &zephyr_discovery.MeshServiceList{
				Items: []zephyr_discovery.MeshService{
					{ObjectMeta: k8s_meta_types.ObjectMeta{ // should NOT be deleted
						Name: fmt.Sprintf("%s-%s-%s", aws.ObjectNamePrefix, *virtualServiceRefs[3].MeshName, meshPlatformName)},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: &zephyr_core_types.ResourceRef{
								Name: "some other non appmesh mesh",
							},
						},
					},
					{ObjectMeta: k8s_meta_types.ObjectMeta{ // should be deleted
						Name: fmt.Sprintf("%s-%s-%s", aws.ObjectNamePrefix, *virtualServiceRefs[0].MeshName, meshPlatformName)},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: &zephyr_core_types.ResourceRef{
								Name: parentMesh.GetName(),
							},
						},
					},
					{ObjectMeta: k8s_meta_types.ObjectMeta{ // should be deleted
						Name: fmt.Sprintf("%s-%s-%s", aws.ObjectNamePrefix, *virtualServiceRefs[1].MeshName, meshPlatformName)},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: &zephyr_core_types.ResourceRef{
								Name: parentMesh.GetName(),
							},
						},
					},
				},
			}
			mockMeshServiceClient.EXPECT().ListMeshService(ctx).Return(existingMeshServices, nil)
			for _, existingMeshService := range existingMeshServices.Items[1:] {
				existingMeshService := existingMeshService
				mockMeshClient.EXPECT().DeleteMesh(ctx, clients.ObjectMetaToObjectKey(existingMeshService.ObjectMeta)).Return(nil)
			}
		}
	}

	It("should reconcile Meshes, MeshWorkloads, and MeshServices", func() {
		meshes := expectReconcileMeshes()
		expectReconcileMeshServices(meshes)
		err := appMeshDiscoveryReconciler.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
	})
})
