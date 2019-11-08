package generator

import (
	"text/template"
)

var GoBinaryDockerfileWithCommonBaseImageTemplate = template.Must(
	template.New("go_binary_dockerfile_with_base_image").
		Funcs(GenBuildFuncs).
		Parse(
			`
{{- $binaryWorkingName := kebab .GoBinaryOutline.BinaryNameBase -}}
# GENERATED FILE - DO NOT EDIT
FROM {{ .Global.BaseImageRepo }}:{{ .Global.BaseImageVersion }}

COPY {{ $binaryWorkingName }}-linux-amd64 /usr/local/bin/{{ $binaryWorkingName }}

ENTRYPOINT ["/usr/local/bin/{{ $binaryWorkingName }}"]
`))
