package clientset_test

import (
	"testing"

	"github.com/solo-io/supergloo/test/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClientset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clientset Suite")
}

var _ = testutils.RegisterCrdsLockingSuite()
