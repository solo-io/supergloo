package appmesh_test

import (
	"context"
	"fmt"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_aws "github.com/solo-io/service-mesh-hub/pkg/aws/parser/mocks"
	settings_utils "github.com/solo-io/service-mesh-hub/pkg/aws/selection"
	mock_settings "github.com/solo-io/service-mesh-hub/pkg/aws/selection/mocks"
	mock_settings2 "github.com/solo-io/service-mesh-hub/pkg/aws/settings/mocks"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	aws4 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	zephyr_discovery_appmesh "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/appmesh"
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
		awsAccountID               string
		appMeshDiscoveryReconciler aws4.RestAPIDiscoveryReconciler
		mockAwsSelector            *mock_settings.MockAwsSelector
		mockSettingsHelperClient   *mock_settings2.MockSettingsHelperClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		awsAccountID = "410461945555"
		mockAppMeshClient = mock_appmesh_clients.NewMockAppMeshAPI(ctrl)
		mockSettingsHelperClient = mock_settings2.NewMockSettingsHelperClient(ctrl)
		mockAwsSelector = mock_settings.NewMockAwsSelector(ctrl)
		appMeshDiscoveryReconciler = zephyr_discovery_appmesh.NewAppMeshDiscoveryReconciler(
			nil,
			func(client client.Client) zephyr_discovery.MeshClient {
				return mockMeshClient
			},
			mockArnParser,
			func(creds *credentials.Credentials, region string) (appmeshiface.AppMeshAPI, error) {
				return mockAppMeshClient, nil
			},
			mockSettingsHelperClient,
			mockAwsSelector,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectReconcileMeshesByRegion = func(region string, selectors []*zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector) {
		page1Input := &appmesh.ListMeshesInput{
			Limit: zephyr_discovery_appmesh.NumItemsPerRequest,
		}
		page2Input := &appmesh.ListMeshesInput{
			Limit:     zephyr_discovery_appmesh.NumItemsPerRequest,
			NextToken: aws2.String("page-2-token"),
		}
		meshRefs := []*appmesh.MeshRef{
			{
				MeshName:  aws2.String("mesh-name-1"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:zephyr_discovery_appmesh:%s:%s:mesh/zephyr_discovery_appmesh-1", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-2"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:zephyr_discovery_appmesh:%s:%s:mesh/zephyr_discovery_appmesh-2", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-3"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:zephyr_discovery_appmesh:%s:%s:mesh/zephyr_discovery_appmesh-3", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-4"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:zephyr_discovery_appmesh:%s:%s:mesh/zephyr_discovery_appmesh-4", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-5"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:zephyr_discovery_appmesh:%s:%s:mesh/zephyr_discovery_appmesh-5", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-6"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:zephyr_discovery_appmesh:%s:%s:mesh/zephyr_discovery_appmesh-6", region, awsAccountID)),
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
		for _, appmeshRef := range page1.Meshes {
			tagsOutput := &appmesh.ListTagsForResourceOutput{
				Tags: []*appmesh.TagRef{{Key: aws2.String("k"), Value: aws2.String("v")}},
			}
			mockAppMeshClient.
				EXPECT().
				ListTagsForResource(&appmesh.ListTagsForResourceInput{ResourceArn: appmeshRef.Arn}).
				Return(tagsOutput, nil)
			mockAwsSelector.EXPECT().AppMeshMatchedBySelectors(appmeshRef, tagsOutput.Tags, selectors).Return(true, nil)
		}
		mockAppMeshClient.EXPECT().ListMeshes(page2Input).Return(page2, nil)
		for _, appmeshRef := range page2.Meshes {
			tagsOutput := &appmesh.ListTagsForResourceOutput{
				Tags: []*appmesh.TagRef{{Key: aws2.String("k"), Value: aws2.String("v")}},
			}
			mockAppMeshClient.
				EXPECT().
				ListTagsForResource(&appmesh.ListTagsForResourceInput{ResourceArn: appmeshRef.Arn}).
				Return(tagsOutput, nil)
			mockAwsSelector.EXPECT().AppMeshMatchedBySelectors(appmeshRef, tagsOutput.Tags, selectors).Return(true, nil)
		}
		for _, meshRef := range meshRefs {
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      metadata.BuildAppMeshName(aws2.StringValue(meshRef.MeshName), region, aws2.StringValue(meshRef.MeshOwner)),
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
						AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
							Name:         *meshRef.MeshName,
							AwsAccountId: awsAccountID,
							Region:       region,
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
			mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
	}

	It("should reconcile Meshes for all regions with empty appmesh_discovery", func() {
		accountID := "accountID"
		region := "region"
		selectors := settings_utils.AwsSelectorsByRegion{
			region: []*zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
				{
					MatcherType: &zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
						Matcher: &zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
							Regions: []string{region},
						},
					},
				},
			},
		}
		settings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId:     accountID,
			MeshDiscovery: &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetMeshDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		expectReconcileMeshesByRegion(region, selectors[region])
		err := appMeshDiscoveryReconciler.Reconcile(ctx, nil, accountID)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile Meshes for all regions with nil appmesh_discovery", func() {
		accountID := "accountID"
		region := "region"
		selectors := settings_utils.AwsSelectorsByRegion{
			region: []*zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
				{
					MatcherType: &zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
						Matcher: &zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
							Regions: []string{region},
						},
					},
				},
			},
		}
		settings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId: accountID,
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetMeshDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		expectReconcileMeshesByRegion(region, selectors[region])
		err := appMeshDiscoveryReconciler.Reconcile(ctx, nil, accountID)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile Meshes for all regions with appmesh_discovery populated with empty resource_selectors", func() {
		accountID := "accountID"
		region := "region"
		selectors := settings_utils.AwsSelectorsByRegion{
			region: []*zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
				{
					MatcherType: &zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
						Matcher: &zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
							Regions: []string{region},
						},
					},
				},
			},
		}
		settings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId: accountID,
			MeshDiscovery: &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{
				ResourceSelectors: []*zephyr_settings_types.SettingsSpec_AwsAccount_ResourceSelector{},
			},
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetMeshDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		expectReconcileMeshesByRegion(region, selectors[region])
		err := appMeshDiscoveryReconciler.Reconcile(ctx, nil, accountID)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not reconcile if mesh_discovery is disabled and clear all Appmeshes from SMH", func() {
		accountID := "accountID"
		settings := &zephyr_settings_types.SettingsSpec_AwsAccount{
			AccountId: "",
			MeshDiscovery: &zephyr_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{
				Disabled: true,
			},
		}
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)

		existingMeshes := &zephyr_discovery.MeshList{
			Items: []zephyr_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh1",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh2",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh3",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(existingMeshes, nil)
		for _, existingMesh := range existingMeshes.Items {
			existingMesh := existingMesh
			mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
		err := appMeshDiscoveryReconciler.Reconcile(ctx, &credentials.Credentials{}, accountID)
		Expect(err).To(BeNil())
	})

	It("should not reconcile if account settings not found and clear all Appmeshes from SMH", func() {
		accountID := "accountID"
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(nil, nil)

		existingMeshes := &zephyr_discovery.MeshList{
			Items: []zephyr_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh1",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh2",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh3",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(existingMeshes, nil)
		for _, existingMesh := range existingMeshes.Items {
			existingMesh := existingMesh
			mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
		err := appMeshDiscoveryReconciler.Reconcile(ctx, &credentials.Credentials{}, accountID)
		Expect(err).To(BeNil())
	})
})
