package traffic_policy_aggregation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAggregation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Traffic Policy Aggregation Suite")
}
