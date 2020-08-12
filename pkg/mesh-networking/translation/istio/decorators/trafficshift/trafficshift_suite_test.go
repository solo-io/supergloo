package trafficshift_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrafficshift(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trafficshift Suite")
}
