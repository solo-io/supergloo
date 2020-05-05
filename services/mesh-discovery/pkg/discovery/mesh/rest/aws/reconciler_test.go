package aws_test

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
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	aws4 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws"
	mock_appmesh_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/aws/appmesh"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Reconciler", func() {
	var (
		ctrl                       *gomock.Controller
		ctx                        context.Context
		mockMeshClient             *mock_core.MockMeshClient
		mockAppMeshClient          *mock_appmesh_clients.MockAppMeshAPI
		mockArnParser              *mock_aws.MockArnParser
		computeTargetName          string
		awsAccountID               string
		appMeshDiscoveryReconciler aws4.RestAPIDiscoveryReconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		computeTargetName = "aws-account-name"
		awsAccountID = "410461945555"
		mockAppMeshClient = mock_appmesh_clients.NewMockAppMeshAPI(ctrl)
		appMeshDiscoveryReconciler = aws.NewAppMeshDiscoveryReconciler(
			mockArnParser,
			mockMeshClient,
			mockAppMeshClient,
			computeTargetName,
			aws4.Region,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectReconcileMeshes = func() {
		region := "us-east-2"
		page1Input := &appmesh.ListMeshesInput{
			Limit: aws.NumItemsPerRequest,
		}
		page2Input := &appmesh.ListMeshesInput{
			Limit:     aws.NumItemsPerRequest,
			NextToken: aws2.String("page-2-token"),
		}
		meshRefs := []*appmesh.MeshRef{
			{
				MeshName:  aws2.String("mesh-name-1"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:appmesh:%s:%s:mesh/appmesh-1", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-2"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:appmesh:%s:%s:mesh/appmesh-2", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-3"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:appmesh:%s:%s:mesh/appmesh-3", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-4"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:appmesh:%s:%s:mesh/appmesh-4", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-5"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:appmesh:%s:%s:mesh/appmesh-5", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-6"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:appmesh:%s:%s:mesh/appmesh-6", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
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
		mockArnParser.EXPECT().ParseAccountID(gomock.Any()).Times(len(meshRefs)).Return(awsAccountID, nil)
		mockAppMeshClient.EXPECT().ListMeshes(page1Input).Return(page1, nil)
		mockAppMeshClient.EXPECT().ListMeshes(page2Input).Return(page2, nil)
		for _, meshRef := range meshRefs {
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      metadata.BuildAppMeshName(aws2.StringValue(meshRef.MeshName), region, aws2.StringValue(meshRef.MeshOwner)),
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
						AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
							Name:         *meshRef.MeshName,
							AwsAccountId: awsAccountID,
							Region:       aws4.Region,
						},
					},
				},
			}
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()}).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			mockMeshClient.
				EXPECT().
				CreateMesh(ctx, mesh).
				Return(nil)
		}
		existingMeshes := &zephyr_discovery.MeshList{
			Items: []zephyr_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{ // should not be deleted
						Name: metadata.BuildAppMeshName(aws2.StringValue(meshRefs[0].MeshName), region, aws2.StringValue(meshRefs[0].MeshOwner)),
					},
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
