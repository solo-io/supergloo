package istio

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIstio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istio mesh ingress plugin Suite")
}

var _ = BeforeSuite(func() {
	kubeClient = fake.NewSimpleClientset()
})
