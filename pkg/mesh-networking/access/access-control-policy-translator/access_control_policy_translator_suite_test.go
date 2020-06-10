package acp_translator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccessControlPolicyTranslator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AccessControlPolicyTranslator Suite")
}
