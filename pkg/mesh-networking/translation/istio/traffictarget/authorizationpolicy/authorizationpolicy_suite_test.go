package authorizationpolicy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthorizationpolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authorizationpolicy Suite")
}
