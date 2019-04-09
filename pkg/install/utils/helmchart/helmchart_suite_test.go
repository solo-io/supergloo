package helmchart_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelmchart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helmchart Suite")
}
