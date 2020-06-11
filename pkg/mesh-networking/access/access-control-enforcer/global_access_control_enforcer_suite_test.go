package access_policy_enforcer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGlobalAccessControlEnforcer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GlobalAccessControlEnforcer Suite")
}
