package upgrade_assets_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAssets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Assets Suite")
}
