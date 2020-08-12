package outlierdetection_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOutlierdetection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Outlierdetection Suite")
}
