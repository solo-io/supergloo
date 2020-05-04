package mc_watcher_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMulticluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multi Cluster Watcher Suite")
}
