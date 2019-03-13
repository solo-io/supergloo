package surveyutils_test

import (
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSurveyutils(t *testing.T) {
	// we allow disabling interactive tests as they appear to fail our current ci environment
	// tracking with https://github.com/solo-io/supergloo/issues/243
	if os.Getenv("SKIP_INTERACTIVE_TESTS") == "1" {
		log.Printf("Skipping interactive tests.")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Surveyutils Suite")
}
