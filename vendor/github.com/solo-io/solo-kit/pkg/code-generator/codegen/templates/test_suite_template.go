package templates

import (
	"text/template"
)

var ProjectTestSuiteTemplate = template.Must(template.New("project_template").Funcs(Funcs).Parse(`package {{ .ProjectConfig.Version }}

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test{{ upper_camel .ProjectConfig.Name }}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{ upper_camel .ProjectConfig.Name }} Suite")
}





`))
