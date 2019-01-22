package templates

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/funcs"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func ProtoFileTemplate(project *model.Project) *template.Template {
	str := `
{{ $File := . -}}

### Package: {{ backtick }}{{ .Package }}{{ backtick }}

{{- if gt (len .SyntaxComments.Detached) 0 }}

{{- range .SyntaxComments.Detached }}  
{{ remove_magic_comments (printf "%v" .) }}

{{- end }}


{{ end }}

{{- if gt (len .Messages) 0 }} 
##### Types:

{{ $linkMessage :=  "- [{{ printfptr \"%v\" .Name }}](#{{ printfptr \"%v\" .Name }}) {{- if (resourceForMessage .) }}** Top-Level Resource**{{ end }}" }}
{{ $linkEnum :=  "- [{{ printfptr \"%v\" .Name }}](#{{ printfptr \"%v\" .Name }})" }}
{{- forEachMessage $File .Messages $linkMessage $linkEnum }}  

{{ end}}

{{- if gt (len .Enums) 0 }} 

##### Enums:

{{ range .Enums}}
	- [{{ printfptr "%v" .Name }}](#{{ printfptr "%v" .Name }})

{{- end}}

{{ end}}

##### Source File: {{ githubLinkForFile "master" .Name }}

{{ $msgLongInfo :=  ` + "`" + `
{{ $Message := . -}}
---
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

{{ if gt (len .Comments.Leading) 0 }} 
{{ remove_magic_comments .Comments.Leading }}
{{- end }}

{{backtick}}{{backtick}}{{backtick}}yaml
{{range .Fields -}}
"{{ printfptr "%v" .Name}}": {{ fieldType . }}
{{end}}
{{backtick}}{{backtick}}{{backtick}}

| Field | Type | Description | Default |
| ----- | ---- | ----------- |----------- | 
{{range .Fields -}}
| {{backtick}}{{ printfptr "%v" .Name }}{{backtick}} | {{linkForField (getFileForMessage $Message) . }} | {{ remove_magic_comments (nobr .Comments.Leading) }} | {{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}

` + "`" + ` }}


{{ $enumLongInfo :=  ` + "`" +
		`
{{ $Enum := . -}}
---
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

{{ if gt (len .Comments.Leading) 0 }} 
{{ remove_magic_comments .Comments.Leading }}
{{- end }}

| Name | Description |
| ----- | ----------- | 
{{range .Values -}}
| {{backtick}}{{ printfptr "%v" .Name }}{{backtick}} | {{ remove_magic_comments (nobr .Comments.Leading) }} |
{{end}}

` + "`" + ` }}

{{- forEachMessage $File .Messages $msgLongInfo $enumLongInfo }}  

{{- range .Enums }}  
### <a name="{{ printfptr "%v" .Name }}">{{ printfptr "%v" .Name }}</a>

Description: {{ remove_magic_comments .Comments.Leading }}

| Name | Description |
| ----- | ----------- | 
{{range .Values -}}
| {{ printfptr "%v" .Name }} | {{ remove_magic_comments (nobr .Comments.Leading) }} |
{{end}}

{{- end }}

<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->
`
	return template.Must(template.New("p").Funcs(funcs.TemplateFuncs(project)).Parse(str))
}
