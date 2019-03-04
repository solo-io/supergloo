package main

import "github.com/solo-io/go-utils/githubutils"

func main() {
	assets := make([]githubutils.ReleaseAssetSpec, 4)
	assets[0] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo-linux-amd64",
		ParentPath: "_output",
		UploadSHA:  true,
	}
	assets[1] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo-darwin-amd64",
		ParentPath: "_output",
		UploadSHA:  true,
	}
	assets[2] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo-windows-amd64.exe",
		ParentPath: "_output",
		UploadSHA:  true,
	}
	assets[3] = githubutils.ReleaseAssetSpec{
		Name:       "supergloo.yaml",
		ParentPath: "install/manifest",
	}
	spec := githubutils.UploadReleaseAssetSpec{
		Owner:             "solo-io",
		Repo:              "supergloo",
		Assets:            assets,
		SkipAlreadyExists: true,
	}
	githubutils.UploadReleaseAssetCli(&spec)
}
