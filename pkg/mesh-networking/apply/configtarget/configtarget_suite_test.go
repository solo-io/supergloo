package configtarget_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfigtarget(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configtarget Suite")
}
