package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/docs/docsgen"
	"github.com/solo-io/go-utils/versionutils"
)

func main() {
	log.Printf("generating necessary site versions")
	repoPath := "./"
	if s := os.Getenv("REPO_PATH"); s != "" {
		repoPath = s
	}
	minVersionStr := "v0.7.1"
	if s := os.Getenv("MIN_SUPPORTED_VERSION"); s != "" {
		minVersionStr = s
	}
	versionJsonPath := "./docs/version.json"
	if s := os.Getenv("VERSION_JSON_PATH"); s != "" {
		versionJsonPath = s
	}
	if err := do(repoPath, versionJsonPath, minVersionStr); err != nil {
		log.Fatal(err)
	}

	log.Printf("finished generating site versions")
}

func do(repoPath, versionJsonPath, minVersionStr string) error {
	repo, err := openRepo(repoPath)
	if err != nil {
		return err
	}
	minVersion, err := versionutils.ParseVersion(minVersionStr)
	if err != nil {
		return err
	}
	versionMeta, err := docsgen.BuildVersionInfo(repo, minVersion)
	if err != nil {
		return err
	}
	f, err := os.Create(versionJsonPath)
	if err != nil {
		return eris.Wrap(err, "unable to create version output file")
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(versionMeta)
}

func openRepo(path string) (*git.Repository, error) {
	if !filepath.IsAbs(path) {
		absRepoPath, err := filepath.Abs(path)
		if err != nil {
			return nil, eris.Wrapf(err, "unable to find absolute path from path %s", path)
		}
		path = absRepoPath
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, eris.Wrapf(err, "unable to open git repository at %s", path)
	}

	return repo, nil
}
