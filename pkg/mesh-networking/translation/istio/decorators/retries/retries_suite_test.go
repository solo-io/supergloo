package retries_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRetries(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Retries Suite")
}
