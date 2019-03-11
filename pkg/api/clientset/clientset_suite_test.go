package clientset_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClientset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clientset Suite")
}
