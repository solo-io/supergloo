package docker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
)

var _ = Describe("Docker image name parser", func() {
	runTest := func(imageName string, expectedParsedImage *docker.Image, expectedErrSubstring string) {
		parsedImage, err := docker.ParseImageName(imageName)

		Expect(parsedImage).To(Equal(expectedParsedImage), "Resulting parsed image metadata should match")

		if expectedErrSubstring == "" {
			Expect(err).To(BeNil(), "Expected no error to be reported")
		} else if err == nil {
			Fail("No error was reported, but we expected: " + expectedErrSubstring)
		} else {
			Expect(err.Error()).To(ContainSubstring(expectedErrSubstring), "Should have reported the expected error")
		}
	}

	DescribeTable("parser", runTest,
		Entry("can parse a non-standardized image name", "consul:1.6.2", &docker.Image{Domain: "docker.io", Path: "library/consul", Tag: "1.6.2"}, ""),
		Entry("handles an image without a tag or digest", "consul", &docker.Image{Domain: "docker.io", Path: "library/consul"}, ""),
		Entry("handles a fully qualified image name", "quay.io/solo-io/discovery:0.20.11", &docker.Image{Domain: "quay.io", Path: "solo-io/discovery", Tag: "0.20.11"}, ""),
	)
})
