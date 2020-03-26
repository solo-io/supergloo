package docgen

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/proto"
	plugin_gogo "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	gendoc "github.com/pseudomuto/protoc-gen-doc"
	"github.com/pseudomuto/protokit"
	"github.com/solo-io/autopilot/codegen/util"
	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/spf13/cobra"
)

var (
	moduleRoot = util.GetModuleRoot()

	cliIndex = `
---
title: "Command-Line Reference"
description: | 
  Detailed descriptions and options for working with the Service Mesh Hub CLI. 
weight: 2
---

This section contains generated reference documentation for the ` + "`" + `Service Mesh Hub` + "`" + ` CLI.

{{% children description="true" %}}

`
	protoDocTemplate  = filepath.Join(moduleRoot, "docs", "proto_docs_template.tmpl")
	apiReferenceIndex = `
---
title: "API Reference"
description: | 
  This section contains the API Specification for the CRDs used by Service Mesh Hub.
weight: 4
---

These docs describe the ` + "`" + `spec` + "`" + ` and ` + "`" + `status` + "`" + ` of the Service Mesh Hub CRDs.

{{% children description="true" %}}

`
)

type CliOptions struct {
	RootCmd   *cobra.Command
	OutputDir string
}

type ProtoOptions struct {
	ProtoRoot string // Will default to vendor_any if empty
	OutputDir string
}

type Options struct {
	Proto    ProtoOptions
	Cli      CliOptions
	DocsRoot string // Will default to docs if empty
}

func Execute(opts Options) error {
	rootDir := filepath.Join(moduleRoot, opts.DocsRoot)
	if err := generateCliReference(rootDir, opts.Cli); err != nil {
		return err
	}
	if err := generateOperatorReference(rootDir, opts.Proto); err != nil {
		return err
	}
	return nil
}

func generateCliReference(root string, opts CliOptions) error {
	// flush directory for idempotence
	cliDocsDir := filepath.Join(root, opts.OutputDir)
	os.RemoveAll(cliDocsDir)
	os.MkdirAll(cliDocsDir, 0755)
	err := clidoc.GenerateCliDocsWithConfig(opts.RootCmd, clidoc.Config{
		OutputDir: cliDocsDir,
	})
	if err != nil {
		return errors.Errorf("error generating docs: %s", err)
	}
	return ioutil.WriteFile(filepath.Join(cliDocsDir, "_index.md"), []byte(cliIndex), 0644)
}

func generateOperatorReference(root string, opts ProtoOptions) error {
	// flush directory for idempotence
	apiDocsDir := filepath.Join(root, opts.OutputDir)
	os.RemoveAll(apiDocsDir)
	os.MkdirAll(apiDocsDir, 0755)

	if opts.ProtoRoot == "" {
		opts.ProtoRoot = filepath.Join(moduleRoot, "vendor_any")
	}
	return generateProtoDocs(opts.ProtoRoot, protoDocTemplate, apiDocsDir, apiReferenceIndex)
}

func generateProtoDocs(protoDir, templateFile, destDir, indexContents string) error {
	tmpDir, err := ioutil.TempDir("", "proto-docs")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	docsTemplate, err := collectDescriptors(protoDir, tmpDir,
		func(file *model.DescriptorWithPath) bool {
			// we only want docs for our protos
			return !strings.HasSuffix(file.GetPackage(), "zephyr.solo.io")
		})
	if err != nil {
		return err
	}

	templateContents, err := ioutil.ReadFile(templateFile)

	tmpl, err := template.New(templateFile).Funcs(templateFuncs).Parse(string(templateContents))
	if err != nil {
		return err
	}

	for _, file := range docsTemplate.Files {
		filename := filepath.Join(destDir, filepath.Base(file.Name))
		filename = strings.TrimSuffix(filename, ".proto") + ".md"
		destFile, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer destFile.Close()
		if err := tmpl.Execute(destFile, file); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filepath.Join(destDir, "_index.md"), []byte(indexContents), 0644)
}

func collectDescriptors(protoDir, outDir string, filter func(file *model.DescriptorWithPath) bool, customImports ...string) (*gendoc.Template, error) {
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

	req := &plugin_gogo.CodeGeneratorRequest{}
	for _, file := range descriptors {
		var added bool
		for _, addedFile := range req.GetFileToGenerate() {
			if addedFile == file.GetName() {
				added = true
			}
		}
		if added {
			continue
		}
		if filter(file) {
			continue
		}
		req.FileToGenerate = append(req.FileToGenerate, file.GetName())
		req.ProtoFile = append(req.ProtoFile, file.FileDescriptorProto)
	}

	// we have to convert the codegen request from a gogo proto to a golang proto
	// because of incompatibility between the solo kit Collector and the
	// psuedomoto/protoc-doc-gen library:
	golangRequest, err := func() (*plugin_go.CodeGeneratorRequest, error) {
		b, err := proto.Marshal(req)
		if err != nil {
			return nil, err
		}
		var golangReq plugin_go.CodeGeneratorRequest
		if err := proto.Unmarshal(b, &golangReq); err != nil {
			return nil, err
		}
		return &golangReq, nil
	}()

	return gendoc.NewTemplate(protokit.ParseCodeGenRequest(golangRequest)), nil
}

var templateFuncs = template.FuncMap{
	"lowerCamel": strcase.ToLowerCamel,
	"replaceNewLine": func(str string) string {
		str = strings.ReplaceAll(str, "\n\n", "<br>")
		return strings.ReplaceAll(str, "\n", " ")
	},
}
