package docsgen

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/iancoleman/strcase"
	gendoc "github.com/pseudomuto/protoc-gen-doc"
	"github.com/pseudomuto/protokit"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
)

var (
	links map[string]string

	protoDocTemplate = filepath.Join(moduleRoot, "docs", "proto_docs_template.tmpl")

	apiReferenceIndex = `
---
title: "API Reference"
description: | 
  This section contains the API Specification for the CRDs used by Gloo Mesh.
weight: 4
---

These docs describe the ` + "`" + `spec` + "`" + ` and ` + "`" + `status` + "`" + ` of the Gloo Mesh CRDs.

{{% children description="true" %}}

`
)

func generateApiDocs(root string, opts ProtoOptions) error {
	// flush directory for idempotence
	apiDocsDir := filepath.Join(root, opts.OutputDir)
	os.RemoveAll(apiDocsDir)
	os.MkdirAll(apiDocsDir, 0755)

	if opts.ProtoRoot == "" {
		opts.ProtoRoot = filepath.Join(moduleRoot, "vendor_any")
	}
	return generateProtoDocs(opts.ProtoRoot, protoDocTemplate, apiDocsDir, apiReferenceIndex)
}

func buildCompleteFilename(destDir string, file *gendoc.File) string {
	filename := filepath.Join(destDir, filepath.Base(file.Name))
	return strings.TrimSuffix(filename, ".proto") + ".md"
}

func generateProtoDocs(protoDir, templateFile, destDir, indexContents string) error {
	tmpDir, err := ioutil.TempDir("", "proto-docs")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	docsTemplate, err := collectDescriptors(protoDir, tmpDir)
	if err != nil {
		return err
	}

	links = collectLinks(destDir, docsTemplate)

	templateContents, err := ioutil.ReadFile(templateFile)

	tmpl, err := template.New(templateFile).Funcs(templateFuncs(links)).Parse(string(templateContents))
	if err != nil {
		return err
	}

	for _, file := range docsTemplate.Files {
		filename := buildCompleteFilename(destDir, file)
		destFile, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer destFile.Close()
		if err := tmpl.Execute(destFile, file); err != nil {
			return err
		}

		tabsetHack(destFile)
	}

	return ioutil.WriteFile(filepath.Join(destDir, "_index.md"), []byte(indexContents), 0644)
}

// in generated markdown files, replace occurrences of the "tabset" shortcode with "tabs"
// because solo-io/hugo-theme-soloio doesn't support "tabset"
// TODO fix code highlighting of tabsets
func tabsetHack(file *os.File) {
	input, err := ioutil.ReadFile(file.Name())
	if err != nil {
		log.Fatalln(err)
	}

	replaced := strings.ReplaceAll(string(input), "tabset", "tabs")

	err = ioutil.WriteFile(file.Name(), []byte(replaced), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

// iterate gendoc template files and construct mapping from proto message name to relative link
func collectLinks(destDir string, template *gendoc.Template) map[string]string {
	links := map[string]string{}
	for _, file := range template.Files {
		filename := filepath.Base(buildCompleteFilename(destDir, file))
		// links consists of "<filename.md>#<message/enumName>"
		for _, msg := range file.Messages {
			if a, ok := links[msg.FullName]; ok && a != msg.FullName {
				log.Printf("warning: found multiple definitions of proto msg %s: %+v", msg.FullName, []string{a, filepath.Base(filename) + "#" + msg.FullName})
			}
			links[msg.FullName] = filepath.Base(filename) + "#" + msg.FullName
		}
		for _, enum := range file.Enums {
			if a, ok := links[enum.FullName]; ok && a != enum.FullName {
				log.Printf("warning: found multiple definitions of proto enum %s: %+v", enum.FullName, []string{a, filepath.Base(filename) + "#" + enum.FullName})
			}
			links[enum.FullName] = filepath.Base(filename) + "#" + enum.FullName
		}
		for _, service := range file.Services {
			if a, ok := links[service.FullName]; ok && a != service.FullName {
				log.Printf("warning: found multiple definitions of proto service %s: %+v", service.FullName, []string{a, filepath.Base(filename) + "#" + service.FullName})
			}
			links[service.FullName] = filepath.Base(filename) + "#" + service.FullName
		}
		for _, extension := range file.Extensions {
			if a, ok := links[extension.FullName]; ok && a != extension.FullName {
				log.Printf("warning: found multiple definitions of proto extension %s: %+v", extension.FullName, []string{a, filepath.Base(filename) + "#" + extension.FullName})
			}
			links[extension.FullName] = filepath.Base(filename) + "#" + extension.FullName
		}
	}
	return links
}

func collectDescriptors(protoDir, outDir string, customImports ...string) (*gendoc.Template, error) {
	descriptors, err := collector.NewCollector(
		customImports,
		[]string{protoDir},
		nil,
		[]string{},
		outDir,
		func(file string) bool {
			return true
		}).CollectDescriptorsFromRoot(protoDir, nil)
	if err != nil {
		return nil, err
	}

	req := &plugin_go.CodeGeneratorRequest{}

	for _, file := range descriptors {
		req.FileToGenerate = append(req.FileToGenerate, file.GetName())
		req.ProtoFile = append(req.ProtoFile, file.FileDescriptorProto)
	}

	return gendoc.NewTemplate(protokit.ParseCodeGenRequest(req)), nil
}

func templateFuncs(links map[string]string) template.FuncMap {
	return template.FuncMap{
		"lowerCamel": strcase.ToLowerCamel,
		"replaceNewLine": func(str string) string {
			str = strings.ReplaceAll(str, "\n\n", "<br>")
			return strings.ReplaceAll(str, "\n", " ")
		},
		"cleanFileName": func(str string) string {
			return filepath.Base(str)
		},
		"link_to_type": func(v interface{}) string {
			switch fieldType := v.(type) {
			case *gendoc.MessageField:
				link, ok := links[fieldType.FullType]
				if ok {
					return fmt.Sprintf("{{< ref \"%s\" >}}", link)
				} else if strings.Contains(fieldType.FullType, ".") {
					fmt.Print(fieldType)
					//panic(fmt.Sprintf("link not found for %s", fieldType))
				}
			}
			return ""
		},
	}
}
