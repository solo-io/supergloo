package generator

import (
	"text/template"
)

// Deprecated: use GoBinaryDockerfileWithCommonBaseImageTemplate
var GoBinaryDockerfileTemplate = template.Must(
	template.New("go_binary_dockerfile").
		Funcs(GenBuildFuncs).
		Parse(
			`
{{- $binaryWorkingName := kebab .GoBinaryOutline.BinaryNameBase -}}
{{ docker_gcloud }}
FROM alpine

RUN apk upgrade --update-cache \
	&& apk add ca-certificates \
	&& rm -rf /var/cache/apk/*

# Install aws-cli
RUN apk -Uuv add groff less python py-pip \
	&& pip install awscli \
	&& apk --purge -v del py-pip\
	&& rm /var/cache/apk/*

COPY {{ $binaryWorkingName }}-linux-amd64 /usr/local/bin/{{ $binaryWorkingName }}

ENTRYPOINT ["/usr/local/bin/{{ $binaryWorkingName }}"]
`))

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
