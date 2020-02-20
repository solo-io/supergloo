{{/*
Expand the name of a container image
*/}}
{{- define "image" -}}
{{ .registry }}/{{ .repository }}:{{ .tag }}
{{- end -}}
