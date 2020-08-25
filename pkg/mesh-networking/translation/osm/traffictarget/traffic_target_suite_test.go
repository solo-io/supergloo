package traffictarget_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrafficTarget(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TrafficTarget Suite")
}
