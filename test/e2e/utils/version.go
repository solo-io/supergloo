package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/solo-io/build/pkg/api/v1"
	"github.com/solo-io/build/pkg/constants"
	"github.com/solo-io/build/pkg/ingest"
)

const defaultSuperglooPath = "github.com/solo-io/supergloo"

// Returns:
// 1. the current BUILD_ID
// 2. the URL where the helm chart for this build is located
// 3. prefix of the repository where the images for this build are located (e.g. quay.io/solo-io)
func GetBuildInformation() (string, string, string, error) {
	sgPath := defaultSuperglooPath
	if customRoot := os.Getenv("PROJECT_ROOT"); customRoot != "" {
		sgPath = customRoot
	}

	projectRoot := filepath.Join(os.Getenv("GOPATH"), "src", sgPath)
	buildConfigFilePath := filepath.Join(projectRoot, constants.DefaultConfigFileName)

	// Get build configuration from build tool
	buildRun, err := ingest.InitializeBuildRun(buildConfigFilePath, &v1.BuildEnvVars{})
	if err != nil {
		return "", "", "", err
	}

	// Set the supergloo version (will be equal to the BUILD_ID env)
	version := buildRun.Config.ComputedBuildVars.Version

	helmRepo := buildRun.Config.ComputedBuildVars.HelmRepository
	helmRepoUrl := strings.Replace(helmRepo, "gs://", "https://storage.googleapis.com/", 1)
	helmRepoUrl = strings.TrimSuffix(helmRepoUrl, "/")

	// Will be similar to: "https://storage.googleapis.com/supergloo-helm-test/charts/supergloo-<BUILD_ID_HERE>.tgz"
	chartUrl := fmt.Sprintf("%s/charts/supergloo-%s.tgz", helmRepoUrl, version)

	imageRepoPrefix := buildRun.Config.ComputedBuildVars.ContainerPrefix

	fmt.Println(fmt.Sprintf("Build info - Version: %s, HelmRepo: %s, ImagePrefix: %s", version, chartUrl, imageRepoPrefix))

	return version, chartUrl, imageRepoPrefix, nil
}
