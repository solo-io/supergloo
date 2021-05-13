package trafficpolicyutils

import (
"testing"

. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"
)

func TestTrafficPolicyUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Traffic Policy Utils Suite")
}
