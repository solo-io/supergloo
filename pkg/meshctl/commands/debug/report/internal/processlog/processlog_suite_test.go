package processlog_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProcessLog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProcessLog Suite")
}
