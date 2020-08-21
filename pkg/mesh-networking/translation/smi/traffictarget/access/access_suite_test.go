package access_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Access Suite")
}
