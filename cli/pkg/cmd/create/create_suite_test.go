package create_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestCreate(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Create Suite")
}
