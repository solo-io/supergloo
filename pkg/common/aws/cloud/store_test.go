package cloud_test

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/cloud"
)

var _ = Describe("Store", func() {
	var (
		ctrl          *gomock.Controller
		awsCloudStore cloud.AwsCloudStore
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		awsCloudStore = cloud.NewAwsCloudStore(func(client appmeshiface.AppMeshAPI) clients.AppmeshClient {
			return nil
		})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should store credentials and their associated regional clients", func() {
		// Store credentials for AWS account
		accountId := "accountId"
		creds := &credentials.Credentials{}
		awsCloudStore.Add(accountId, creds)
		// Retrieve credentials
		region1 := "region1"
		awsCloudForRegion1a, err := awsCloudStore.Get(accountId, region1)
		Expect(err).ToNot(HaveOccurred())
		awsCloudForRegion1b, err := awsCloudStore.Get(accountId, region1)
		Expect(err).ToNot(HaveOccurred())
		Expect(awsCloudForRegion1a).To(BeIdenticalTo(awsCloudForRegion1b))
	})

	It("should store and remove credentials and their associated regional clients", func() {
		// Store credentials for AWS account
		accountId := "accountId"
		creds := &credentials.Credentials{}
		awsCloudStore.Add(accountId, creds)
		// Initialize credentials
		region1 := "region1"
		awsCloudForRegion1a, err := awsCloudStore.Get(accountId, region1)
		Expect(err).ToNot(HaveOccurred())
		region2 := "region2"
		awsCloudForRegion2a, err := awsCloudStore.Get(accountId, region2)
		Expect(err).ToNot(HaveOccurred())
		// Retrieve existing clients
		awsCloudForRegion1b, err := awsCloudStore.Get(accountId, region1)
		Expect(err).ToNot(HaveOccurred())
		awsCloudForRegion2b, err := awsCloudStore.Get(accountId, region2)
		Expect(err).ToNot(HaveOccurred())
		Expect(awsCloudForRegion1a).To(BeIdenticalTo(awsCloudForRegion1b))
		Expect(awsCloudForRegion2a).To(BeIdenticalTo(awsCloudForRegion2b))
		// Remove account
		awsCloudStore.Remove(accountId)
		_, err = awsCloudStore.Get(accountId, region1)
		Expect(err).To(testutils.HaveInErrorChain(cloud.AwsCredsNotFound(err, accountId)))
	})
})
