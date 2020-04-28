package aws_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws"
)

var _ = Describe("ArnParser", func() {
	var arnParser = aws.NewArnParser()

	It("should parse AWS account ID", func() {
		accountID := "123456"
		arnString := fmt.Sprintf("arn:aws:iam::%s:role/iamserviceaccount-role", accountID)
		id, err := arnParser.ParseAccountID(arnString)
		Expect(err).To(BeNil())
		Expect(id).To(Equal(accountID))
	})

	It("should throw error for invalid ARN", func() {
		arnString := "invalid"
		_, err := arnParser.ParseAccountID(arnString)
		Expect(err).To(testutils.HaveInErrorChain(aws.ARNParseError(err, arnString)))
	})
})
