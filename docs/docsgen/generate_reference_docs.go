package docsgen

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/proto"
	plugin_gogo "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/google/go-github/github"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	gendoc "github.com/pseudomuto/protoc-gen-doc"
	"github.com/pseudomuto/protokit"
	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/skv2/codegen/util"
	"github.com/solo-io/solo-kit/pkg/code-generator/collector"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	moduleRoot = util.GetModuleRoot()

	cliIndex = `
---
title: "Command-Line Reference"
description: | 
  Detailed descriptions and options for working with the Gloo Mesh CLI. 
weight: 2
---

This section contains generated reference documentation for the ` + "`" + `Gloo Mesh` + "`" + ` CLI.

`
	protoDocTemplate  = filepath.Join(moduleRoot, "docs", "proto_docs_template.tmpl")
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
	changelogIndex = `
---
title: "Changelog"
description: | 
  Section containing Changelogs for Gloo Mesh
weight: 6
---

Included in the sections below are Changelog for both Gloo Mesh OSS and Gloo Mesh Enterprise.

{{% children description="true" %}}

`
	changelogTypes = `
---
title: Changelog Entry Types
weight: 90
description: Explanation of the entry types used in our changelogs
---

You will find several different kinds of changelog entries:

- **Dependency Bumps**

A notice about a dependency in the project that had its version bumped in this release. Be sure to check for any
"**Breaking Change**" entries accompanying a dependency bump. For example, a ` + "`gloo-mesh`" + `
version bump in Gloo Mesh Enterprise may mean a change to a proto API.

- **Breaking Changes**

A notice of a non-backwards-compatible change to some API. This can include things like a changed
proto format, a change to the Helm chart, and other breakages. Occasionally a breaking change
may mean that the process to upgrade the product is slightly different; in that case, we will be sure
to specify in the changelog how the break must be handled.

- **Helm Changes**

A notice of a change to our Helm chart. One of these entries does not by itself signify a breaking
change to the Helm chart; you will find an accompanying "**Breaking Change**" entry in the release
notes if that is the case.

- **New Features**

A description of a new feature that has been implemented in this release.

- **Fixes**

A description of a bug that was resolved in this release.

`
	changelogTmpl = `
---
title: {{.Name}} 
description: |
  Changelogs for {{.Name}}
weight: {{.Weight}}
---

{{.Body}}

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

type ChangelogConfig struct {
	Name string
	Repo string
	Path string
}

type ChangelogOptions struct {
	Repos     []ChangelogConfig
	OutputDir string
}

type Options struct {
	Proto     ProtoOptions
	Cli       CliOptions
	Changelog ChangelogOptions
	DocsRoot  string // Will default to docs if empty
}

func Execute(opts Options) error {
	rootDir := filepath.Join(moduleRoot, opts.DocsRoot)
	if err := generateCliReference(rootDir, opts.Cli); err != nil {
		return err
	}
	if err := generateOperatorReference(rootDir, opts.Proto); err != nil {
		return err
	}
	if err := generateChangelog(rootDir, opts.Changelog); err != nil {
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
			return !strings.HasSuffix(file.GetPackage(), "mesh.gloo.solo.io")
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
	// pseudomuto/protoc-gen-doc library:
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

func generateChangelog(root string, opts ChangelogOptions) error {
	fmt.Println("version: " + os.Getenv("TAGGED_VERSION"))
	fmt.Println("release: " + os.Getenv("RELEASE"))
	if strings.ToLower(os.Getenv("RELEASE")) != `"true"` {
		fmt.Println("skip?")
		return nil
	}

	// flush directory for idempotence
	changelogDir := filepath.Join(root, opts.OutputDir)
	os.RemoveAll(changelogDir)
	os.MkdirAll(changelogDir, 0755)
	type tplParams struct {
		ChangelogConfig
		Body   string
		Weight int
	}
	tmpl := template.Must(template.New("changelog").Parse(changelogTmpl))
	for i, cfg := range opts.Repos {
		body, err := generateChangelogMD(cfg.Repo)
		if err != nil {
			return err
		}
		if err := func() error {
			f, err := os.Create(filepath.Join(changelogDir, cfg.Path+".md"))
			if err != nil {
				return err
			}
			defer f.Close()
			return tmpl.Execute(f, tplParams{ChangelogConfig: cfg, Body: body, Weight: 7 + i})
		}(); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(
		filepath.Join(changelogDir, "changelog_types.md"), []byte(changelogTypes), 0644,
	); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(changelogDir, "_index.md"), []byte(changelogIndex), 0644)
}

var (
	githubClient            *github.Client
	MissingGithubTokenError = errors.New("Must set GITHUB_TOKEN environment variable")
)

func getGithubClient() (*github.Client, error) {
	if githubClient != nil {
		return githubClient, nil
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, MissingGithubTokenError
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient = github.NewClient(tc)
	return githubClient, nil
}

func generateChangelogMD(repo string) (string, error) {
	client, err := getGithubClient()
	if err != nil {
		return "", err
	}

	releases, _, err := client.Repositories.ListReleases(context.Background(), "solo-io", repo, &github.ListOptions{Page: 0, PerPage: 1000000})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, release := range releases {
		sb.WriteString("### " + *release.TagName + "\n\n")
		sb.WriteString(*release.Body + "\n")
	}

	return sb.String(), nil
}

var templateFuncs = template.FuncMap{
	"lowerCamel": strcase.ToLowerCamel,
	"replaceNewLine": func(str string) string {
		str = strings.ReplaceAll(str, "\n\n", "<br>")
		return strings.ReplaceAll(str, "\n", " ")
	},
	"cleanFileName": func(str string) string {
		return filepath.Base(str)
	},
}
