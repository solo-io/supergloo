package aws_utils_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
)

var _ = Describe("ArnParser", func() {
	var arnParser = aws_utils.NewArnParser()

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
		Expect(err).To(testutils.HaveInErrorChain(aws_utils.ARNParseError(err, arnString)))
	})
})
