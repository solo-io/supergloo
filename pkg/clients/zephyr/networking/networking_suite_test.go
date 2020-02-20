package zephyr_networking_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNetworking(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Networking Suite")
}
