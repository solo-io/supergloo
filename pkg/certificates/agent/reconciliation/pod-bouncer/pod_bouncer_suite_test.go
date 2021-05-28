package podbouncer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPodBouncer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PodBouncer Suite")
}
