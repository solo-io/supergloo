package plugins_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPlugins(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugins Suite")
}
