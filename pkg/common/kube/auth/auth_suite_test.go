package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Auth Suite")
}
