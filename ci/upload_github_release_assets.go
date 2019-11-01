package main

import "github.com/solo-io/go-utils/githubutils"

func main() {
	assets := make([]githubutils.ReleaseAssetSpec, 1)
	assets[0] = githubutils.ReleaseAssetSpec{
		Name:       "sm-marketplace.yaml",
		ParentPath: "install",
	}
	spec := githubutils.UploadReleaseAssetSpec{
		Owner:             "solo-io",
		Repo:              "sm-marketplace",
		Assets:            assets,
		SkipAlreadyExists: true,
	}
	githubutils.UploadReleaseAssetCli(&spec)
}
