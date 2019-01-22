package templates

import (
	"text/template"
)

var ResourceGroupSnapshotTemplate = template.Must(template.New("resource_group_snapshot").Funcs(Funcs).Parse(
	`package {{ .Project.ProjectConfig.Version }}

import (
	{{ .Imports }}
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/hashutils"
	"go.uber.org/zap"
)

type {{ .GoName }}Snapshot struct {
{{- range .Resources}}
{{- if .ClusterScoped }}
	{{ upper_camel .PluralName }} {{ .ImportPrefix }}{{ .Name }}List
{{- else }}
	{{ upper_camel .PluralName }} {{ .ImportPrefix }}{{ upper_camel .PluralName }}ByNamespace
{{- end }}
{{- end}}
}

func (s {{ .GoName }}Snapshot) Clone() {{ .GoName }}Snapshot {
	return {{ .GoName }}Snapshot{
{{- range .Resources}}
		{{ upper_camel .PluralName }}: s.{{ upper_camel .PluralName }}.Clone(),
{{- end}}
	}
}

func (s {{ .GoName }}Snapshot) Hash() uint64 {
	return hashutils.HashAll(
{{- range .Resources}}
		s.hash{{ upper_camel .PluralName }}(),
{{- end}}
	)
}

{{- $ResourceGroup := . }}
{{- range .Resources }}

func (s {{ $ResourceGroup.GoName }}Snapshot) hash{{ upper_camel .PluralName }}() uint64 {
{{- if .ClusterScoped }}
	return hashutils.HashAll(s.{{ upper_camel .PluralName }}.AsInterfaces()...)
{{- else }}
	return hashutils.HashAll(s.{{ upper_camel .PluralName }}.List().AsInterfaces()...)
{{- end }}
}
{{- end}}

func (s {{ .GoName }}Snapshot) HashFields() []zap.Field {
	var fields []zap.Field

{{- range .Resources}}
	fields = append(fields, zap.Uint64("{{ lower_camel .PluralName }}", s.hash{{ upper_camel .PluralName }}() ))
{{- end}}

	return append(fields, zap.Uint64("snapshotHash",  s.Hash()))
}

`))
