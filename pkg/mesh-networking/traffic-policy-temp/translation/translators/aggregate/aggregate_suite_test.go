package mesh_translation_aggregate_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAggregate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aggregate Suite")
}
