package docsgen

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
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
	cliIndex   = `
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
)

var changelogTpl = template.Must(template.New("changelog").Parse(`---
title: {{.Name}} 
description: |
  Changelogs for {{.Name}}
weight: {{.Weight}}
---
{{.Body}}
`))

type CliOptions struct {
	RootCmd   *cobra.Command
	OutputDir string
}

type ProtoOptions struct {
	ProtoRoot string // Will default to vendor_any if empty
	OutputDir string
}

type ChangelogConfig struct {
	Name  string
	Repo  string
	Fname string
}

type ChangelogOptions struct {
	OutputDir  string
	OtherRepos []ChangelogConfig
}

type Options struct {
	Proto     ProtoOptions
	Cli       CliOptions
	Changelog ChangelogOptions
	DocsRoot  string // Will default to docs if empty
}

func Execute(opts Options) error {
	rootDir := filepath.Join(moduleRoot, opts.DocsRoot)
	if os.Getenv("ONLY_CHANGELOG") == "" {
		if err := generateCliReference(rootDir, opts.Cli); err != nil {
			return err
		}
		if err := generateApiDocs(rootDir, opts.Proto); err != nil {
			return err
		}
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
	if os.Getenv("SKIP_CHANGELOG_GENERATION") != "" {
		return nil
	}
	// flush directory for idempotence
	changelogDir := filepath.Join(root, opts.OutputDir)
	os.RemoveAll(changelogDir)
	os.MkdirAll(changelogDir, 0755)
	version, err := getGitVersion()
	if err != nil {
		return err
	}

	// Generate community changelog
	if err := generateChangelogMd(
		"gloo-mesh", "Gloo Mesh Community", filepath.Join(changelogDir, "community.md"), version, 7,
	); err != nil {
		return err
	}

	// Generate changelog for other repos
	for i, cfg := range opts.OtherRepos {
		if err := generateChangelogMd(
			cfg.Repo, cfg.Name, filepath.Join(changelogDir, cfg.Fname+".md"), "", 8+i,
		); err != nil {
			return err
		}
	}

	// Write the reference
	if err := ioutil.WriteFile(
		filepath.Join(changelogDir, "changelog_types.md"), []byte(changelogTypes), 0644,
	); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(changelogDir, "_index.md"), []byte(changelogIndex), 0644)
}

func generateChangelogMd(repo, name, path, cutoffVersion string, weight int) error {
	body, err := buildChangelogBody(repo, cutoffVersion)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return changelogTpl.Execute(f, struct {
		Name   string
		Body   string
		Weight int
	}{Name: name, Body: body, Weight: weight})
}

func buildChangelogBody(repo, cutoffVersion string) (string, error) {
	client, err := getGitHubClient()
	if err != nil {
		return "", err
	}
	releases, _, err := client.Repositories.ListReleases(
		context.Background(),
		"solo-io", repo,
		&github.ListOptions{Page: 0, PerPage: 1000000},
	)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	passedCutoff := cutoffVersion == "" // If no cutoff version provided, include all versions
	for _, release := range releases {
		// Check if we've found the cutoff, and if we haven't, continue
		if !passedCutoff && release.GetTagName() == cutoffVersion {
			passedCutoff = true
		} else if !passedCutoff {
			continue
		}

		// Only write releases that have changelogs in their bodies
		if release.GetBody() != "" {
			sb.WriteString("### " + release.GetTagName() + "\n\n")
			sb.WriteString(release.GetBody() + "\n")
		}
	}
	return sb.String(), nil
}

var ghClient *github.Client

func getGitHubClient() (*github.Client, error) {
	if ghClient != nil {
		return ghClient, nil
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("must set GITHUB_TOKEN environment variable")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	ghClient = github.NewClient(tc)
	return ghClient, nil
}

// getGitVersion finds the current version via the checked out git reference
func getGitVersion() (string, error) {
	branch, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}
	if len(branch) > 0 {
		return "", nil
	}
	version, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(version)), nil
}
