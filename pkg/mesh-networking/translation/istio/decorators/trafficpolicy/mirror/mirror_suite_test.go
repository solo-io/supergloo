package mirror_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMirror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mirror Suite")
}
