package aws_utils_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	mock_kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	"github.com/solo-io/go-utils/testutils"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser"
	mock_aws "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser/mocks"
	mock_multicluster "github.com/solo-io/service-mesh-hub/test/mocks/smh/clients"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AwsAccountIdFetcher", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		mockArnParser       *mock_aws.MockArnParser
		mockConfigMapClient *mock_kubernetes_core.MockConfigMapClient
		mockMcClient        *mock_multicluster.MockClient
		awsAccountIdFetcher aws_utils.AwsAccountIdFetcher
		accountID           = aws_utils.AwsAccountId("111122223333")
		roleARN             = fmt.Sprintf("arn:aws:iam::%s:role/role-name", accountID)
		configMap           = &k8s_core_types.ConfigMap{
			Data: map[string]string{
				"mapRoles": fmt.Sprintf(`
- groups:
  - system:bootstrappers
  - system:nodes
  rolearn: %s
  username: system:node:{{EC2PrivateDNSName}}
`, roleARN),
			},
		}
		clusterName = "cluster-name"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockConfigMapClient = mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		mockMcClient = mock_multicluster.NewMockClient(ctrl)
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		awsAccountIdFetcher = aws_utils.NewAwsAccountIdFetcher(
			mockArnParser,
			func(client client.Client) k8s_core.ConfigMapClient {
				return mockConfigMapClient
			},
			mockMcClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectGetClusterClient = func() {
		// Doesn't matter what is returned here because the configMapClientFactory will always return mockConfigMapClient
		mockMcClient.EXPECT().Cluster(clusterName).Return(nil, nil)
	}

	It("should fetch AWS account ID from ConfigMap", func() {
		expectGetClusterClient()
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, aws_utils.AwsAuthConfigMapKey).Return(configMap, nil)
		mockArnParser.EXPECT().ParseAccountID(roleARN).Return(string(accountID), nil)
		id, err := awsAccountIdFetcher.GetEksAccountId(ctx, clusterName)
		Expect(err).To(BeNil())
		Expect(id).To(Equal(accountID))
	})

	It("should throw error for invalid ARN", func() {
		expectGetClusterClient()
		testErr := eris.New("test")
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, aws_utils.AwsAuthConfigMapKey).Return(configMap, nil)
		mockArnParser.EXPECT().ParseAccountID(roleARN).Return("", testErr)
		id, err := awsAccountIdFetcher.GetEksAccountId(ctx, clusterName)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(id).To(BeEmpty())
	})

	It("should return empty if ConfigMap not found", func() {
		expectGetClusterClient()
		notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, aws_utils.AwsAuthConfigMapKey).Return(nil, notFoundErr)
		id, err := awsAccountIdFetcher.GetEksAccountId(ctx, clusterName)
		Expect(err).To(BeNil())
		Expect(id).To(BeEmpty())
	})
})
