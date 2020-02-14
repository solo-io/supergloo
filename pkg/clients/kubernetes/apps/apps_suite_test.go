package kubernetes_apps_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apps Suite")
}
