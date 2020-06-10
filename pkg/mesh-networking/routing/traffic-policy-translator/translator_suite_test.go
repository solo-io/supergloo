package traffic_policy_translator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrafficPolicyTranslator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "trafficPolicyTranslator Suite")
}
