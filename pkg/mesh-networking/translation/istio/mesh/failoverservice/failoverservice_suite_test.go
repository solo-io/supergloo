package failoverservice_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFailoverservice(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Failoverservice Suite")
}
