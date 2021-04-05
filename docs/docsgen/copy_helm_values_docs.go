package docsgen

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/rotisserie/eris"
)

var (
	helmDocsDir = "content/reference/helm"

	enterpriseNetworkingHelmValueDocPath = "enterprise-networking/codegen/helm/enterprise_networking_helm_values_reference.md"
	enterpriseAgentHelmValueDocPath      = "enterprise-networking/codegen/helm/enterprise_agent_helm_values_reference.md"

	fileMapping = map[string]string{
		enterpriseNetworkingHelmValueDocPath: "%s/%s/enterprise_networking.md",
		enterpriseAgentHelmValueDocPath:      "%s/%s/enterprise_agent.md",
	}

	helmValuesIndex = `
---
title: "%s"
description: Reference for Helm values. 
weight: 2
---

The following pages provide reference documentation for Helm values for the various Gloo Mesh
components. These components include:

1. **Open source Gloo Mesh**: the OSS version of Gloo Mesh
2. **Enterprise Networking (enterprise only)**: the management plane of Gloo Mesh Enterprise, deployed on the management cluster
3. **Enterprise Agent (enterprise only)**: the agent of Gloo Mesh Enterprise, deployed on each managed cluster
4. **RBAC Webhook (enterprise only)**: the Kubernetes webhook that enforces Gloo Mesh Enterprise's role-based API
5. **Gloo Mesh UI (enterprise only)**: the UI for Gloo Mesh Enterprise

Note that when providing Helm values for the bundled Gloo Mesh Enterprise chart 
(located at https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise),
values for each subchart must be prefixed accordingly:

1. Values for the RBAC Webhook must be prefixed with "rbac-webhook".
2. Values for Enterprise Networking must be prefixed with "enterprise-networking".
3. Values for the Gloo Mesh UI must be prefixed with "gloo-mesh-ui".


{{%% children description="true" %%}}
`
)

func copyHelmValuesDocsFromEnterprise(client *github.Client, rootDir string) error {
	// flush directory for idempotence
	helmDocsDir := filepath.Join(rootDir, helmDocsDir)
	os.RemoveAll(helmDocsDir)
	os.MkdirAll(helmDocsDir, 0755)

	if err := createFileIfNotExists(helmDocsDir+"/"+"_index.md", fmt.Sprintf(helmValuesIndex, "Helm Values Reference")); err != nil {
		return eris.Errorf("error creating Helm values index file: %v", err)
	}

	// include Helm values docs for all versions > v1.0.0-beta14
	releases, _, err := client.Repositories.ListReleases(
		context.Background(),
		GithubOrg,
		GlooMeshEnterpriseRepoName,
		&github.ListOptions{Page: 0, PerPage: 1000000},
	)
	if err != nil {
		return eris.Errorf("error listing releases: %v", err)
	}
	var tags []string
	for _, release := range releases {
		if release.GetTagName() == "v1.0.0-beta14" {
			break
		}
		tags = append(tags, release.GetTagName())
	}

	for _, tag := range tags {
		if err := os.Mkdir(helmDocsDir+"/"+tag, os.ModePerm); err != nil {
			return eris.Errorf("error creating Helm docs directories: %v", err)
		}

		if err := createFileIfNotExists(helmDocsDir+"/"+tag+"/"+"_index.md", fmt.Sprintf(helmValuesIndex, tag)); err != nil {
			return eris.Errorf("error creating Helm values index file: %v", err)
		}

		for src, dest := range fileMapping {
			dest = fmt.Sprintf(dest, helmDocsDir, tag)
			if err := copyHelmValuesDocs(client, GithubOrg, GlooMeshEnterpriseRepoName, tag, src, dest); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyHelmValuesDocs(client *github.Client, org, repo, tag, path, destinationFile string) error {
	contents, _, _, err := client.Repositories.GetContents(context.Background(), org, repo, path, &github.RepositoryContentGetOptions{
		Ref: tag,
	})
	if err != nil {
		return eris.Errorf("error fetching Helm values doc: %v", err)
	}

	decodedContents, err := base64.StdEncoding.DecodeString(*contents.Content)
	if err != nil {
		return eris.Errorf("error fetching Helm values doc: %v", err)
	}

	return createFileIfNotExists(destinationFile, string(decodedContents))
}

// create file with contents, create the file if it doesn't exist
func createFileIfNotExists(fname, contents string) error {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return eris.Errorf("error copying Helm values reference doc: %v", err)
	}

	if _, err = f.Write([]byte(contents)); err != nil {
		return err
	}

	return nil
}
