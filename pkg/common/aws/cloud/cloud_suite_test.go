package cloud_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCloud(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloud Suite")
}
