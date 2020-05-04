package aws_test

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	aws2 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	rest3 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest"
	mock_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws/clients/appmesh/mocks"
	mock_rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/mocks"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CredsHandler", func() {
	var (
		ctrl                           *gomock.Controller
		ctx                            context.Context
		mockAppMeshClientFactory       *mock_aws.MockAppMeshClientFactory
		mockRestAPIDiscoveryReconciler *mock_rest_api.MockRestAPIDiscoveryReconciler
		appMeshClient                  *appmesh.AppMesh
		awsCredsHandler                aws2.AwsCredsHandler
		secret                         *k8s_core_types.Secret
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		appMeshClient = &appmesh.AppMesh{}
		mockAppMeshClientFactory = mock_aws.NewMockAppMeshClientFactory(ctrl)
		mockRestAPIDiscoveryReconciler = mock_rest_api.NewMockRestAPIDiscoveryReconciler(ctrl)
		awsCredsHandler = aws2.NewAwsAPIHandler(
			mockAppMeshClientFactory,
			func(_ string, _ appmeshiface.AppMeshAPI, _ string) rest3.RestAPIDiscoveryReconciler {
				return mockRestAPIDiscoveryReconciler
			})
		secret = &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "secret-name",
				Namespace: "service-mesh-hub",
			},
			Type: aws_creds.AWSSecretType,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should ignore non-AWS secrets when compute target added", func() {
		err := awsCredsHandler.ComputeTargetAdded(ctx, &k8s_core_types.Secret{Type: k8s_core_types.SecretTypeOpaque})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should ignore non-AWS secrets when compute target removed", func() {
		err := awsCredsHandler.ComputeTargetRemoved(ctx, &k8s_core_types.Secret{Type: k8s_core_types.SecretTypeOpaque})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should handle new API registration", func() {
		mockAppMeshClientFactory.EXPECT().Build(secret, aws2.Region).Return(appMeshClient, nil)
		//mockRestAPIDiscoveryReconciler.EXPECT().Reconcile(gomock.Any())
		err := awsCredsHandler.ComputeTargetAdded(ctx, secret)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should handle new API deregistration", func() {
		// register the API first for cancelFunc map entry
		mockAppMeshClientFactory.EXPECT().Build(secret, aws2.Region).Return(appMeshClient, nil)
		err := awsCredsHandler.ComputeTargetAdded(ctx, secret)
		Expect(err).ToNot(HaveOccurred())

		err = awsCredsHandler.ComputeTargetRemoved(ctx, secret)
		Expect(err).ToNot(HaveOccurred())
	})
})
