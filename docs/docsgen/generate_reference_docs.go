package docsgen

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/skv2/codegen/util"
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
	Generate  bool
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
	if err := generateApiDocs(rootDir, opts.Proto); err != nil {
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

func generateChangelog(root string, opts ChangelogOptions) error {
	if !opts.Generate {
		fmt.Println("skipping changelog generation, pass --changelog to enable")
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
		body, err := generateChangelogMD(cfg.Repo, cfg.Version)
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

func generateChangelogMD(repo, version string) (string, error) {
	client, err := getGithubClient()
	if err != nil {
		return "", err
	}

	releases, _, err := client.Repositories.ListReleases(context.Background(), "solo-io", repo, &github.ListOptions{Page: 0, PerPage: 1000000})
	if err != nil {
		return "", err
	}

	var (
		sb           strings.Builder
		foundVersion = version == "" // Include all versions if none specified
	)
	for _, release := range releases {
		// Do not include versions after the provided version
		if !foundVersion && release.GetTagName() == version {
			foundVersion = true
		} else if !foundVersion {
			continue
		}

		sb.WriteString("### " + *release.TagName + "\n\n")
		sb.WriteString(*release.Body + "\n")
	}

	return sb.String(), nil
}
