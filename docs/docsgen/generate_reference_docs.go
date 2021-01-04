package docsgen

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/skv2/codegen/util"
	"github.com/spf13/cobra"
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
	Name    string
	Repo    string
	Path    string
	Version string
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

func Execute(ctx context.Context, opts Options) error {
	rootDir := filepath.Join(moduleRoot, opts.DocsRoot)
	if err := generateCliReference(rootDir, opts.Cli); err != nil {
		return err
	}
	if err := generateApiDocs(rootDir, opts.Proto); err != nil {
		return err
	}
	if err := generateChangelog(ctx, rootDir, opts.Changelog); err != nil {
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

func generateChangelog(ctx context.Context, root string, opts ChangelogOptions) error {
	if os.Getenv("SKIP_CHANGELOG_GENERATION") != "" {
		fmt.Println("skipping changelog generation")
		return nil
	}

	// flush directory for idempotence
	changelogDir := filepath.Join(root, opts.OutputDir)
	os.RemoveAll(changelogDir)
	os.MkdirAll(changelogDir, 0755)

	changelog, err := readChangelog(ctx, "gloo-mesh") // TODO(ryantking): Un-hardcode once we have multiple repos
	if err != nil {
		return eris.Wrap(err, "building changelog from local files")
	}

	type tplParams struct {
		ChangelogConfig
		Body   string
		Weight int
	}
	tmpl := template.Must(template.New("changelog").Parse(changelogTmpl))
	for i, cfg := range opts.Repos {
		if err := func() error {
			f, err := os.Create(filepath.Join(changelogDir, cfg.Path+".md"))
			if err != nil {
				return err
			}
			defer f.Close()
			return tmpl.Execute(f, tplParams{ChangelogConfig: cfg, Body: changelog, Weight: 7 + i})
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

// returns the changelog represented as a map of versions to the rendered body
func readChangelog(ctx context.Context, repo string) (string, error) {
	var buf bytes.Buffer
	if err := changelogutils.GenerateChangelogFromLocalDirectory(
		ctx,
		"./",
		"solo-io",
		repo,
		"changelog/",
		&buf,
	); err != nil {
		return "", nil
	}

	return buf.String(), nil
}
