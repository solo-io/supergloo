package upstream_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEdit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Edit Suite")
}
