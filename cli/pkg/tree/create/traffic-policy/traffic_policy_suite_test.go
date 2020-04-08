package traffic_policy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrafficpolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trafficpolicy Suite")
}
