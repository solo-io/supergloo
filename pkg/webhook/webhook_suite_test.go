package webhook

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestPatch(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Sidecar Injection Webhook Suite")
}
