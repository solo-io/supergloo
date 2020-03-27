package exploration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/explore/exploration"
)

var _ = Describe("Resource name parser", func() {
	var resourceNameAsserter = func(resourceName string, parsedName *exploration.FullyQualifiedKubeResource, expectedErr error) {
		output, err := exploration.ParseResourceName(resourceName)
		if expectedErr == nil {
			Expect(err).NotTo(HaveOccurred(), "No error expected from this test case")
			Expect(output).To(Equal(parsedName))
		} else {
			Expect(err).NotTo(BeNil(), "Expected an error")
			Expect(err).To(testutils.HaveInErrorChain(expectedErr))
			Expect(output).To(BeNil())
		}
	}
	DescribeTable("Resource name parser", resourceNameAsserter,
		Entry("can parse expected resource names", "resource-name.resource-namespace.cluster-name", &exploration.FullyQualifiedKubeResource{
			Name:        "resource-name",
			Namespace:   "resource-namespace",
			ClusterName: "cluster-name",
		}, nil),
		Entry("fails on just a name by itself", "resource-name", nil, exploration.InvalidResourceName("resource-name")),
		Entry("fails on just a name and namespace", "resource-name.resource-namespace", nil, exploration.InvalidResourceName("resource-name.resource-namespace")),
		Entry("fails on invalid characters", "resource-name.invalid-n&mespace.cluster-name", nil, exploration.InvalidResourceName("resource-name.invalid-n&mespace.cluster-name")),
	)
})
