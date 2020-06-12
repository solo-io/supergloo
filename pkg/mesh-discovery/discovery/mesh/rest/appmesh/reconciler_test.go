package appmesh_test

import (
	"context"
	"fmt"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	mock_appmesh "github.com/solo-io/service-mesh-hub/pkg/common/aws/clients/mocks"
	cloud2 "github.com/solo-io/service-mesh-hub/pkg/common/aws/cloud"
	mock_cloud2 "github.com/solo-io/service-mesh-hub/pkg/common/aws/cloud/mocks"
	mock_aws "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser/mocks"
	settings_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/selection"
	mock_settings "github.com/solo-io/service-mesh-hub/pkg/common/aws/selection/mocks"
	mock_settings2 "github.com/solo-io/service-mesh-hub/pkg/common/aws/settings/mocks"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	aws4 "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws"
	smh_discovery_appmesh "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/rest/appmesh"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
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
		mockAppmeshClient          *mock_appmesh.MockAppmeshClient
		mockArnParser              *mock_aws.MockArnParser
		awsAccountID               string
		appMeshDiscoveryReconciler aws4.RestAPIDiscoveryReconciler
		mockAwsSelector            *mock_settings.MockAwsSelector
		mockAwsCloudStore          *mock_cloud2.MockAwsCloudStore
		mockSettingsHelperClient   *mock_settings2.MockSettingsHelperClient
		region                     = "region"
		accountId                  = "accountID"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		awsAccountID = "410461945555"
		mockAppmeshClient = mock_appmesh.NewMockAppmeshClient(ctrl)
		mockSettingsHelperClient = mock_settings2.NewMockSettingsHelperClient(ctrl)
		mockAwsSelector = mock_settings.NewMockAwsSelector(ctrl)
		mockAwsCloudStore = mock_cloud2.NewMockAwsCloudStore(ctrl)
		appMeshDiscoveryReconciler = smh_discovery_appmesh.NewAppMeshDiscoveryReconciler(
			nil,
			func(client client.Client) smh_discovery.MeshClient {
				return mockMeshClient
			},
			mockArnParser,
			mockSettingsHelperClient,
			mockAwsSelector,
			mockAwsCloudStore,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectGetAwsCloud = func() {
		awsCloud := &cloud2.AwsCloud{
			Appmesh: mockAppmeshClient,
		}
		mockAwsCloudStore.EXPECT().Get(accountId, region).Return(awsCloud, nil)
	}

	var expectReconcileMeshesByRegion = func(region string, selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector) {
		page1Input := &appmesh.ListMeshesInput{
			Limit: smh_discovery_appmesh.NumItemsPerRequest,
		}
		page2Input := &appmesh.ListMeshesInput{
			Limit:     smh_discovery_appmesh.NumItemsPerRequest,
			NextToken: aws2.String("page-2-token"),
		}
		meshRefs := []*appmesh.MeshRef{
			{
				MeshName:  aws2.String("mesh-name-1"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:smh_discovery_appmesh:%s:%s:mesh/smh_discovery_appmesh-1", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-2"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:smh_discovery_appmesh:%s:%s:mesh/smh_discovery_appmesh-2", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-3"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:smh_discovery_appmesh:%s:%s:mesh/smh_discovery_appmesh-3", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-4"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:smh_discovery_appmesh:%s:%s:mesh/smh_discovery_appmesh-4", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-5"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:smh_discovery_appmesh:%s:%s:mesh/smh_discovery_appmesh-5", region, awsAccountID)),
				MeshOwner: aws2.String(awsAccountID),
			},
			{
				MeshName:  aws2.String("mesh-name-6"),
				Arn:       aws2.String(fmt.Sprintf("arn:aws:smh_discovery_appmesh:%s:%s:mesh/smh_discovery_appmesh-6", region, awsAccountID)),
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
		mockAppmeshClient.EXPECT().ListMeshes(page1Input).Return(page1, nil)
		for _, appmeshRef := range page1.Meshes {
			tagsOutput := &appmesh.ListTagsForResourceOutput{
				Tags: []*appmesh.TagRef{{Key: aws2.String("k"), Value: aws2.String("v")}},
			}
			mockAppmeshClient.
				EXPECT().
				ListTagsForResource(&appmesh.ListTagsForResourceInput{ResourceArn: appmeshRef.Arn}).
				Return(tagsOutput, nil)
			mockAwsSelector.EXPECT().AppMeshMatchedBySelectors(appmeshRef, tagsOutput.Tags, selectors).Return(true, nil)
		}
		mockAppmeshClient.EXPECT().ListMeshes(page2Input).Return(page2, nil)
		for _, appmeshRef := range page2.Meshes {
			tagsOutput := &appmesh.ListTagsForResourceOutput{
				Tags: []*appmesh.TagRef{{Key: aws2.String("k"), Value: aws2.String("v")}},
			}
			mockAppmeshClient.
				EXPECT().
				ListTagsForResource(&appmesh.ListTagsForResourceInput{ResourceArn: appmeshRef.Arn}).
				Return(tagsOutput, nil)
			mockAwsSelector.EXPECT().AppMeshMatchedBySelectors(appmeshRef, tagsOutput.Tags, selectors).Return(true, nil)
		}
		for _, meshRef := range meshRefs {
			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      metadata.BuildAppMeshName(aws2.StringValue(meshRef.MeshName), region, aws2.StringValue(meshRef.MeshOwner)),
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshSpec{
					MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
						AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
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
		existingMeshes := &smh_discovery.MeshList{
			Items: []smh_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{ // should not be deleted
						Name: metadata.BuildAppMeshName(aws2.StringValue(meshRefs[0].MeshName), region, aws2.StringValue(meshRefs[0].MeshOwner)),
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "non-existent-1"},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}}}, // should be deleted
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "non-existent-2"},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}}}, // should be deleted
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(existingMeshes, nil)
		for _, existingMesh := range existingMeshes.Items[1:] {
			existingMesh := existingMesh
			mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
	}

	It("should reconcile Meshes for all regions with empty appmesh_discovery", func() {
		expectGetAwsCloud()
		selectors := settings_utils.AwsSelectorsByRegion{
			region: []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
				{
					MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
						Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
							Regions: []string{region},
						},
					},
				},
			},
		}
		settings := &smh_settings_types.SettingsSpec_AwsAccount{
			AccountId:     accountId,
			MeshDiscovery: &smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetMeshDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountId).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		expectReconcileMeshesByRegion(region, selectors[region])
		err := appMeshDiscoveryReconciler.Reconcile(ctx, accountId)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile Meshes for all regions with nil appmesh_discovery", func() {
		expectGetAwsCloud()
		selectors := settings_utils.AwsSelectorsByRegion{
			region: []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
				{
					MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
						Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
							Regions: []string{region},
						},
					},
				},
			},
		}
		settings := &smh_settings_types.SettingsSpec_AwsAccount{
			AccountId: accountId,
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetMeshDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountId).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		expectReconcileMeshesByRegion(region, selectors[region])
		err := appMeshDiscoveryReconciler.Reconcile(ctx, accountId)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile Meshes for all regions with appmesh_discovery populated with empty resource_selectors", func() {
		expectGetAwsCloud()
		selectors := settings_utils.AwsSelectorsByRegion{
			region: []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{
				{
					MatcherType: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher_{
						Matcher: &smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector_Matcher{
							Regions: []string{region},
						},
					},
				},
			},
		}
		settings := &smh_settings_types.SettingsSpec_AwsAccount{
			AccountId: accountId,
			MeshDiscovery: &smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{
				ResourceSelectors: []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{},
			},
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetMeshDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountId).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		expectReconcileMeshesByRegion(region, selectors[region])
		err := appMeshDiscoveryReconciler.Reconcile(ctx, accountId)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not reconcile if mesh_discovery is disabled and clear all Appmeshes from SMH", func() {
		accountID := "accountID"
		settings := &smh_settings_types.SettingsSpec_AwsAccount{
			AccountId: "",
			MeshDiscovery: &smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{
				Disabled: true,
			},
		}
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)

		existingMeshes := &smh_discovery.MeshList{
			Items: []smh_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh2",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh3",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(existingMeshes, nil)
		for _, existingMesh := range existingMeshes.Items {
			existingMesh := existingMesh
			mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
		err := appMeshDiscoveryReconciler.Reconcile(ctx, accountID)
		Expect(err).To(BeNil())
	})

	It("should not reconcile if account settings not found and clear all Appmeshes from SMH", func() {
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountId).Return(nil, nil)

		existingMeshes := &smh_discovery.MeshList{
			Items: []smh_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh2",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh3",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{}}},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(existingMeshes, nil)
		for _, existingMesh := range existingMeshes.Items {
			existingMesh := existingMesh
			mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(existingMesh.ObjectMeta)).Return(nil)
		}
		err := appMeshDiscoveryReconciler.Reconcile(ctx, accountId)
		Expect(err).To(BeNil())
	})
})
