package detector_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDetector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Detector Suite")
}
