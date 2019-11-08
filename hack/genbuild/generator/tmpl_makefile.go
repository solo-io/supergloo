package generator

import (
	"text/template"
)

var assumptions = `
Dockerfile is in the same dir as the main function's file
`
var GoBinaryMakefileBuildTemplate = template.Must(
	template.New("go_binary_makefile_build").
		Funcs(GenBuildFuncs).
		Parse(`
#----------------------------------------------------------------------------------
# {{ .BinaryNameBase }}
# Generated with args: {{ jsoner . }}
#----------------------------------------------------------------------------------
{{ $binaryDir := makefileArg .BinaryNameBase "DIR" }} 
{{- $sources := makefileArg .BinaryNameBase "SOURCES" }} 
{{- $binaryDir }}={{ .BinaryDir }}
{{ $sources }}=$(shell find $({{ $binaryDir }}) -name "*.go" | grep -v test | grep -v generated.go)

$(OUTPUT_DIR)/{{ .BinaryNameBase }}-linux-amd64: $({{ $sources }})
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $({{ $binaryDir }})/main.go

.PHONY: {{ .BinaryNameBase }}
{{ .BinaryNameBase }}: $(OUTPUT_DIR)/{{ .BinaryNameBase }}-linux-amd64

$(OUTPUT_DIR)/Dockerfile.{{ .BinaryNameBase }}: $({{ $binaryDir }})/Dockerfile
	cp $< $@

.PHONY: {{ .BinaryNameBase }}-docker
{{ .BinaryNameBase }}-docker: $(OUTPUT_DIR)/.{{ .BinaryNameBase }}-docker

$(OUTPUT_DIR)/.{{ .BinaryNameBase }}-docker: $(OUTPUT_DIR)/{{ .BinaryNameBase }}-linux-amd64 $(OUTPUT_DIR)/Dockerfile.{{ .BinaryNameBase }}
	docker build -t quay.io/solo-io/{{ .ImageName }}:$(VERSION) $(call get_test_tag_option,{{ .BinaryNameBase }}) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.{{ .BinaryNameBase }}
	touch $@
`))
