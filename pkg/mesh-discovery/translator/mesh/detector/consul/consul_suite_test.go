package consul_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConsul(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consul Suite")
}
