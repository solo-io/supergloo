package eks_test

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/clients/cluster-registration"
	mock_clients "github.com/solo-io/service-mesh-hub/pkg/clients/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	compute_target_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	discovery_eks "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s-cluster/rest/eks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
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
		mockKubeClusterClient         *mock_core.MockKubernetesClusterClient
		mockEksClient                 *mock_cloud.MockEksClient
		mockEksConfigBuilder          *mock_discovery.MockEksConfigBuilder
		mockClusterRegistrationClient *mock_clients.MockClusterRegistrationClient
		eksReconciler                 compute_target_aws.EksDiscoveryReconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeClusterClient = mock_core.NewMockKubernetesClusterClient(ctrl)
		mockEksClient = mock_cloud.NewMockEksClient(ctrl)
		mockEksConfigBuilder = mock_discovery.NewMockEksConfigBuilder(ctrl)
		mockClusterRegistrationClient = mock_clients.NewMockClusterRegistrationClient(ctrl)
		eksReconciler = discovery_eks.NewEksDiscoveryReconciler(
			mockKubeClusterClient,
			func(creds *credentials.Credentials, region string) (cloud.EksClient, error) {
				return mockEksClient, nil
			},
			func(eksClient cloud.EksClient) discovery.EksConfigBuilder {
				return mockEksConfigBuilder
			},
			mockClusterRegistrationClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectFetchEksClustersOnAWS = func(region string) (sets.String, map[string]string) {
		nextToken := "next-token"
		awsCluster1Name := "cluster1"
		smhCluster1Name := metadata.BuildEksClusterName(awsCluster1Name, region)
		awsCluster2Name := "cluster2"
		smhCluster2Name := metadata.BuildEksClusterName(awsCluster2Name, region)
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
		mockEksClient.
			EXPECT().
			ListClusters(ctx, &eks.ListClustersInput{
				MaxResults: aws.Int64(discovery_eks.MaxResults),
				NextToken:  aws.String(nextToken),
			}).
			Return(&eks.ListClustersOutput{
				Clusters: []*string{aws.String(awsCluster2Name)},
			}, nil)
		return sets.NewString(smhCluster1Name, smhCluster2Name), smhToAwsClusterNames
	}

	var expectFetchEksClustersOnSMH = func() sets.String {
		clusterList := &zephyr_discovery.KubernetesClusterList{
			Items: []zephyr_discovery.KubernetesCluster{
				{ObjectMeta: v1.ObjectMeta{Name: "cluster1"}},
			},
		}
		mockKubeClusterClient.
			EXPECT().
			ListKubernetesCluster(ctx, client.MatchingLabels(
				map[string]string{constants.DISCOVERED_BY: discovery_eks.ReconcilerDiscoverySource}),
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
			env.GetWriteNamespace(),
			"",
			discovery_eks.ReconcilerDiscoverySource,
			cluster_registration.ClusterRegisterOpts{},
		).Return(nil)
	}

	It("should reconcile", func() {
		region := "region"
		clustersOnAWS, smhToAwsClusterNames := expectFetchEksClustersOnAWS(region)
		clustersOnSMH := expectFetchEksClustersOnSMH()
		clustersToRegister := clustersOnAWS.Difference(clustersOnSMH)
		for _, clusterToRegister := range clustersToRegister.List() {
			expectRegisterCluster(smhToAwsClusterNames[clusterToRegister], clusterToRegister)
		}
		err := eksReconciler.Reconcile(ctx, &credentials.Credentials{}, region)
		Expect(err).To(BeNil())
	})
})
