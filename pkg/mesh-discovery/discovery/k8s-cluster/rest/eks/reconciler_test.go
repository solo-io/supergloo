package eks_test

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_settings_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	settings_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/selection"
	mock_settings "github.com/solo-io/service-mesh-hub/pkg/common/aws/selection/mocks"
	mock_settings2 "github.com/solo-io/service-mesh-hub/pkg/common/aws/settings/mocks"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/common/cluster-registration"
	mock_registration "github.com/solo-io/service-mesh-hub/pkg/common/cluster-registration/mocks"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	compute_target_aws "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws"
	discovery_eks "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/k8s-cluster/rest/eks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/multicluster/discovery"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
	mock_cloud "github.com/solo-io/skv2/pkg/multicluster/discovery/cloud/mocks"
	mock_discovery "github.com/solo-io/skv2/pkg/multicluster/discovery/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Reconciler", func() {
	var (
		ctrl                          *gomock.Controller
		ctx                           context.Context
		accountID                     = "accountid"
		mockKubeClusterClient         *mock_core.MockKubernetesClusterClient
		mockEksClient                 *mock_cloud.MockEksClient
		mockEksConfigBuilder          *mock_discovery.MockEksConfigBuilder
		mockClusterRegistrationClient *mock_registration.MockClusterRegistrationClient
		mockAwsSelector               *mock_settings.MockAwsSelector
		mockSettingsHelperClient      *mock_settings2.MockSettingsHelperClient
		eksReconciler                 compute_target_aws.EksDiscoveryReconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeClusterClient = mock_core.NewMockKubernetesClusterClient(ctrl)
		mockEksClient = mock_cloud.NewMockEksClient(ctrl)
		mockEksConfigBuilder = mock_discovery.NewMockEksConfigBuilder(ctrl)
		mockClusterRegistrationClient = mock_registration.NewMockClusterRegistrationClient(ctrl)
		mockSettingsHelperClient = mock_settings2.NewMockSettingsHelperClient(ctrl)
		mockAwsSelector = mock_settings.NewMockAwsSelector(ctrl)
		eksReconciler = discovery_eks.NewEksDiscoveryReconciler(
			mockKubeClusterClient,
			func(creds *credentials.Credentials, region string) (cloud.EksClient, error) {
				return mockEksClient, nil
			},
			func(eksClient cloud.EksClient) discovery.EksConfigBuilder {
				return mockEksConfigBuilder
			},
			mockClusterRegistrationClient,
			mockSettingsHelperClient,
			mockAwsSelector,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectFetchEksClustersOnAWS = func(
		region string,
		selectors []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector,
	) (sets.String, map[string]string) {
		nextToken := "next-token"
		awsCluster1Name := "cluster1"
		smhCluster1Name := metadata.BuildEksKubernetesClusterName(awsCluster1Name, region)
		awsCluster2Name := "cluster2"
		smhCluster2Name := metadata.BuildEksKubernetesClusterName(awsCluster2Name, region)
		smhToAwsClusterNames := map[string]string{
			smhCluster1Name: awsCluster1Name,
			smhCluster2Name: awsCluster2Name,
		}
		mockEksClient.
			EXPECT().
			ListClusters(ctx, &eks.ListClustersInput{
				MaxResults: aws.Int64(discovery_eks.MaxResults),
			}).
			Return(&eks.ListClustersOutput{
				Clusters:  []*string{aws.String(awsCluster1Name)},
				NextToken: aws.String(nextToken),
			}, nil)
		eksCluster1 := &eks.Cluster{}
		mockEksClient.EXPECT().DescribeCluster(ctx, awsCluster1Name).Return(eksCluster1, nil)
		mockAwsSelector.EXPECT().EKSMatchedBySelectors(eksCluster1, selectors).Return(true, nil)
		mockEksClient.
			EXPECT().
			ListClusters(ctx, &eks.ListClustersInput{
				MaxResults: aws.Int64(discovery_eks.MaxResults),
				NextToken:  aws.String(nextToken),
			}).
			Return(&eks.ListClustersOutput{
				Clusters: []*string{aws.String(awsCluster2Name)},
			}, nil)
		eksCluster2 := &eks.Cluster{}
		mockEksClient.EXPECT().DescribeCluster(ctx, awsCluster2Name).Return(eksCluster2, nil)
		mockAwsSelector.EXPECT().EKSMatchedBySelectors(eksCluster2, selectors).Return(true, nil)
		return sets.NewString(smhCluster1Name, smhCluster2Name), smhToAwsClusterNames
	}

	var expectFetchEksClustersOnSMH = func() sets.String {
		clusterList := &smh_discovery.KubernetesClusterList{
			Items: []smh_discovery.KubernetesCluster{
				{ObjectMeta: v1.ObjectMeta{Name: "cluster1"}},
			},
		}
		mockKubeClusterClient.
			EXPECT().
			ListKubernetesCluster(ctx, client.MatchingLabels(
				map[string]string{kube.DISCOVERED_BY: discovery_eks.EKSClusterDiscoveryLabel}),
			).Return(clusterList, nil)
		return sets.NewString(clusterList.Items[0].GetName())
	}

	var expectRegisterCluster = func(awsName, smhName string) {
		eksCluster := &eks.Cluster{}
		configForEksCluster := &clientcmd.DirectClientConfig{}
		mockEksClient.EXPECT().DescribeCluster(ctx, awsName).Return(eksCluster, nil)
		mockEksConfigBuilder.EXPECT().ConfigForCluster(ctx, eksCluster).Return(configForEksCluster, nil)
		mockClusterRegistrationClient.EXPECT().Register(
			ctx,
			configForEksCluster,
			smhName,
			container_runtime.GetWriteNamespace(),
			"",
			discovery_eks.EKSClusterDiscoveryLabel,
			cluster_registration.ClusterRegisterOpts{},
		).Return(nil)
	}

	It("should reconcile for all regions with empty eks_discovery object", func() {
		region := "region"
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
			AccountId:    accountID,
			EksDiscovery: &smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{},
		}
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetEksDiscovery()).Return(true)
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		clustersOnAWS, smhToAwsClusterNames := expectFetchEksClustersOnAWS(region, selectors[region])
		clustersOnSMH := expectFetchEksClustersOnSMH()
		clustersToRegister := clustersOnAWS.Difference(clustersOnSMH)
		for _, clusterToRegister := range clustersToRegister.List() {
			expectRegisterCluster(smhToAwsClusterNames[clusterToRegister], clusterToRegister)
		}
		err := eksReconciler.Reconcile(ctx, &credentials.Credentials{}, accountID)
		Expect(err).To(BeNil())
	})

	It("should reconcile for all regions with eks_discovery populated with empty resource_selectors", func() {
		region := "region"
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
			AccountId: accountID,
			EksDiscovery: &smh_settings_types.SettingsSpec_AwsAccount_DiscoverySelector{
				ResourceSelectors: []*smh_settings_types.SettingsSpec_AwsAccount_ResourceSelector{},
			},
		}
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetEksDiscovery()).Return(false)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		clustersOnAWS, smhToAwsClusterNames := expectFetchEksClustersOnAWS(region, selectors[region])
		clustersOnSMH := expectFetchEksClustersOnSMH()
		clustersToRegister := clustersOnAWS.Difference(clustersOnSMH)
		for _, clusterToRegister := range clustersToRegister.List() {
			expectRegisterCluster(smhToAwsClusterNames[clusterToRegister], clusterToRegister)
		}
		err := eksReconciler.Reconcile(ctx, &credentials.Credentials{}, accountID)
		Expect(err).To(BeNil())
	})

	It("should reconcile all regions if eks_discovery is nil", func() {
		region := "region"
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
		settings := &smh_settings_types.SettingsSpec_AwsAccount{}
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(settings, nil)
		mockAwsSelector.EXPECT().IsDiscoverAll(settings.GetEksDiscovery()).Return(true)
		mockAwsSelector.EXPECT().AwsSelectorsForAllRegions().Return(selectors)

		clustersOnAWS, smhToAwsClusterNames := expectFetchEksClustersOnAWS(region, selectors[region])
		clustersOnSMH := expectFetchEksClustersOnSMH()
		clustersToRegister := clustersOnAWS.Difference(clustersOnSMH)
		for _, clusterToRegister := range clustersToRegister.List() {
			expectRegisterCluster(smhToAwsClusterNames[clusterToRegister], clusterToRegister)
		}
		err := eksReconciler.Reconcile(ctx, &credentials.Credentials{}, accountID)
		Expect(err).To(BeNil())
	})

	It("should not reconcile if account settings not found", func() {
		mockSettingsHelperClient.EXPECT().GetAWSSettingsForAccount(ctx, accountID).Return(nil, nil)

		expectFetchEksClustersOnSMH()
		err := eksReconciler.Reconcile(ctx, &credentials.Credentials{}, accountID)
		Expect(err).To(BeNil())
	})
})
