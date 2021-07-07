package docsgen

import (
	"context"
	changelogdocutils "github.com/solo-io/go-utils/changeloggenutils"
	"github.com/solo-io/go-utils/githubutils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/skv2/codegen/util"
	"github.com/spf13/cobra"
)

const (
	GithubOrg                  = "solo-io"
	GlooMeshEnterpriseRepoName = "gloo-mesh-enterprise"
	GlooMeshRepoName           = "gloo-mesh"
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
	if !checkEnvVariable("ONLY_CHANGELOG") {
		if err := generateCliReference(rootDir, opts.Cli); err != nil {
			return err
		}
		if err := generateApiDocs(rootDir, opts.Proto); err != nil {
			return err
		}
	}

	client, err := githubutils.GetClient(context.Background())
	if err != nil {
		return eris.Errorf("error initializing Github client: %v", err)
	}

	// fetch Helm values docs from Gloo Mesh Enterprise
	if err = copyHelmValuesDocsForAllCharts(client, rootDir); err != nil {
		return err
	}

	// generate changelog documentation
	if err := generateChangelog(client, rootDir, opts.Changelog); err != nil {
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

func generateChangelog(client *github.Client, root string, opts ChangelogOptions) error {
	if checkEnvVariable("SKIP_CHANGELOG_GENERATION") {
		return nil
	}
	// Generate community changelog
	generator := changelogdocutils.NewMinorReleaseGroupedChangelogGenerator(changelogdocutils.Options{
		MainRepo: "gloo-mesh",
		RepoOwner: "solo-io",
	}, client)
	out, err := generator.GenerateJSON(context.Background())
	if err != nil {
		return err
	}
	f, err := os.Create("content/static/content/community_changelog.json")
	if err != nil {
		return err
	}
	_, err = f.WriteString(out)
	if err != nil {
		return eris.Wrap(err, "error writing changelog")
	}
	err = f.Close()
	if err != nil {
		return err
	}

	// Generate changelog for gloo mesh enterprise
	ghToken, err := githubutils.GetGithubToken()
	if err != nil {
		return err
	}
	depFn, err := changelogdocutils.GetOSDependencyFunc("solo-io", "gloo-mesh-enterprise", "gloo-mesh", ghToken)
	if err != nil {
		return eris.Wrap(err, "unable to generate dependency function between gloo-mesh-enterprise and gloo-mesh community")
	}
	mergedOpts := changelogdocutils.Options{
		MainRepo: "gloo-mesh-enterprise",
		DependentRepo: "gloo-mesh",
		RepoOwner: "solo-io",
	}
	mergedGenerator := changelogdocutils.NewMergedReleaseGeneratorWithDepFn(mergedOpts, client, depFn)
	out, err = mergedGenerator.GenerateJSON(context.Background())
	if err != nil {
		return eris.Wrap(err,"error generating merged enterprise and open source changelog")
	}

	f, err = os.Create("content/static/content/enterprise_changelog.json")
	if err != nil {
		return err
	}
	_, err = f.WriteString(out)
	if err != nil {
		return eris.Wrap(err, "error writing changelog")
	}
	err = f.Close()
	return err
}

func generateChangelogMd(client *github.Client, repo, name, path, cutoffVersion string, weight int) error {
	body, err := buildChangelogBody(client, repo, cutoffVersion)
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

func buildChangelogBody(client *github.Client, repo, cutoffVersion string) (string, error) {
	releases, _, err := client.Repositories.ListReleases(
		context.Background(),
		GithubOrg, repo,
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

func checkEnvVariable(key string) bool {
	val := os.Getenv(key)
	if val == "" {
		return false
	}
	b, err := strconv.ParseBool(val)
	return err != nil || b // treat set env variables as true
}
