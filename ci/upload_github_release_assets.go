package main

import (
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/logger"
	"github.com/solo-io/go-utils/pkgmgmtutils"
)

func main() {
	const buildDir = "_output"
	const repoOwner = "solo-io"
	const repoName = "supergloo"

	assets := make([]githubutils.ReleaseAssetSpec, 4)
	assets[0] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo-cli-linux-amd64",
		ParentPath: buildDir,
		UploadSHA:  true,
	}
	assets[1] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo-cli-darwin-amd64",
		ParentPath: buildDir,
		UploadSHA:  true,
	}
	assets[2] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo-cli-windows-amd64.exe",
		ParentPath: buildDir,
		UploadSHA:  true,
	}
	assets[3] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo.yaml",
		ParentPath: "install/manifest",
	}

	spec := githubutils.UploadReleaseAssetSpec{
		Owner:             repoOwner,
		Repo:              repoName,
		Assets:            assets,
		SkipAlreadyExists: true,
	}
	githubutils.UploadReleaseAssetCli(&spec)

	fOpts := []pkgmgmtutils.FormulaOptions{
		{
			Name:           "homebrew-tap/supergloo",
			FormulaName:    "supergloo",
			Path:           "Formula/supergloo.rb",
			RepoOwner:      repoOwner,      // Make change in this repo owner
			RepoName:       "homebrew-tap", //   expects this repo is forked from PRRepoOwner if PRRepoOwner != RepoOwner
			PRRepoOwner:    repoOwner,      // Make PR to this repo owner
			PRRepoName:     "homebrew-tap", //   and this repo
			PRBranch:       "master",       //   and merge into this branch
			PRDescription:  "",
			PRCommitName:   "Solo-io Bot",
			PRCommitEmail:  "bot@solo.io",
			VersionRegex:   `version\s*"([0-9.]+)"`,
			DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
			LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,
		},
	}

	// Update package manager install formulas
	status, err := pkgmgmtutils.UpdateFormulas(repoOwner, repoName, buildDir, fOpts)
	if err != nil {
		logger.Fatalf("Error trying to update package manager formulas. Error was: %s", err.Error())
	}
	for _, s := range status {
		if !s.Updated {
			if s.Err != nil {
				logger.Fatalf("Error while trying to update formula %s. Error was: %s", s.Name, s.Err.Error())
			} else {
				logger.Fatalf("Error while trying to update formula %s. Error was nil", s.Name) // Shouldn't happen; really bad if it does
			}
		}
		if s.Err != nil {
			if s.Err == pkgmgmtutils.ErrAlreadyUpdated {
				logger.Warnf("Formula %s was updated externally, so no updates applied during this release", s.Name)
			} else {
				logger.Fatalf("Error updating Formula %s. Error was: %s", s.Name, s.Err.Error())
			}
		}
	}
}
