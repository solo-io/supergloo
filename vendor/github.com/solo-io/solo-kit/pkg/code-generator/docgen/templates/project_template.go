package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/funcs"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProjectDocsRootTemplate(project *model.Project) *template.Template {
	return template.Must(template.New("pf").Funcs(funcs.TemplateFuncs(project)).Parse(`

### API Reference for {{ .ProjectConfig.Title}}

API Version: ` + "`{{ .ProjectConfig.Name }}.{{ .ProjectConfig.Version }}`" + `

{{ .ProjectConfig.Description }}

### API Resources:
{{- range .Resources}}
- {{linkForResource . }}
{{- end}}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`))
}
