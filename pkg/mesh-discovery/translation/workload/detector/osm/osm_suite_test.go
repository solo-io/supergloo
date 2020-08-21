package osm_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOsm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Osm Suite")
}
