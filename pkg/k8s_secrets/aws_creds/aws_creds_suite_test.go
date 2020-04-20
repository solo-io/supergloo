package aws_creds_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAwsCreds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AwsCreds Suite")
}
