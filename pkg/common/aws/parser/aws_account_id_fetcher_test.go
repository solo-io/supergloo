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
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockConfigMapClient = mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		mockArnParser = mock_aws.NewMockArnParser(ctrl)
		awsAccountIdFetcher = aws_utils.NewAwsAccountIdFetcher(
			mockArnParser,
			func(client client.Client) k8s_core.ConfigMapClient {
				return mockConfigMapClient
			},
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should fetch AWS account ID from ConfigMap", func() {
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, aws_utils.AwsAuthConfigMapKey).Return(configMap, nil)
		mockArnParser.EXPECT().ParseAccountID(roleARN).Return(string(accountID), nil)
		id, err := awsAccountIdFetcher.GetEksAccountId(ctx, nil)
		Expect(err).To(BeNil())
		Expect(id).To(Equal(accountID))
	})

	It("should throw error for invalid ARN", func() {
		testErr := eris.New("test")
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, aws_utils.AwsAuthConfigMapKey).Return(configMap, nil)
		mockArnParser.EXPECT().ParseAccountID(roleARN).Return("", testErr)
		id, err := awsAccountIdFetcher.GetEksAccountId(ctx, nil)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(id).To(BeEmpty())
	})

	It("should return empty if ConfigMap not found", func() {
		notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
		mockConfigMapClient.EXPECT().GetConfigMap(ctx, aws_utils.AwsAuthConfigMapKey).Return(nil, notFoundErr)
		id, err := awsAccountIdFetcher.GetEksAccountId(ctx, nil)
		Expect(err).To(BeNil())
		Expect(id).To(BeEmpty())
	})
})
