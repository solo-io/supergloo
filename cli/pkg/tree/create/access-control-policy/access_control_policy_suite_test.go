package access_control_policy

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccesscontrolpolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Access control policy Suite")
}
