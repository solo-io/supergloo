package consul_test

import (
	"errors"

	"github.com/solo-io/service-mesh-hub/pkg/common/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/consul"
	kubev1 "k8s.io/api/core/v1"
)

var _ = Describe("Consul Connect Installation Finder", func() {
	testErr := errors.New("")

	runTestCase := func(container kubev1.Container, shouldBeConsul bool, expectedErrMessageSubstring string) {
		connectInstallationFinder := consul.NewConsulConnectInstallationScanner(docker.NewImageNameParser())

		isConsulConnect, err := connectInstallationFinder.IsConsulConnect(container)
		Expect(isConsulConnect).To(Equal(shouldBeConsul), "Should have correctly determined whether connect is deployed")

		if expectedErrMessageSubstring == "" {
			Expect(err).To(BeNil(), "Expected no error to be reported")
		} else if err == nil {
			Fail("No error was reported, but we expected: " + expectedErrMessageSubstring)
		} else {
			Expect(err.Error()).To(ContainSubstring(expectedErrMessageSubstring), "Should have reported the expected error")
		}
	}

	DescribeTable("installation finder", runTestCase,
		Entry("can detect a basic deployment", consulDeployment().Spec.Template.Spec.Containers[0], true, ""),
		Entry("does not detect consul when it is not present", kubev1.Container{Name: "not-consul", Image: "test-image:0.6.9"}, false, ""),
		Entry(
			"responds with false and reports an error when the image name is malformed",
			kubev1.Container{
				Name:  "bad-image-name",
				Image: "bad-image-name:4.2.0:whoops:oops",
			},
			false,
			consul.InvalidImageFormatError(testErr, "bad-image-name:4.2.0:whoops:oops").Error(),
		),
	)
})
