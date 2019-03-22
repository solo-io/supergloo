package gloo_test

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGloo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gloo Suite")
}

var _ = BeforeSuite(func() {
	kubeClient = fake.NewSimpleClientset()
})
