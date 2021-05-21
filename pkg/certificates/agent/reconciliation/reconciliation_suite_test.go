package reconciliation

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestReconciliation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reconciliation Suite")
}
