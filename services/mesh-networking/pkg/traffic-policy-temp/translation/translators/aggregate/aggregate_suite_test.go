package mesh_translation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAggregate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aggregate Suite")
}
