package faultinjection_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFaultinjection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Faultinjection Suite")
}
