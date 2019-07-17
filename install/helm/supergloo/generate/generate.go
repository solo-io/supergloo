package generate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	glooGenerate "github.com/solo-io/gloo/install/helm/gloo/generate"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/supergloo/pkg/constants"
)

const (
	DefaultValues = "install/helm/supergloo/values-defaults.yaml"
	ValuesOutput  = "install/helm/supergloo/values.yaml"
	ChartTemplate = "install/helm/supergloo/Chart-template.yaml"
	ChartOutput   = "install/helm/supergloo/Chart.yaml"

	gopkgToml  = "Gopkg.toml"
	constraint = "constraint"
)

var rootPrefix = ""

func Run(version, imageTag, imageRepoPrefix string) error {
	glooVersion, err := getOsGlooVersion(rootPrefix)
	if err != nil {
		return err
	}

	if err := generateValuesYaml(imageTag, imageRepoPrefix, glooVersion); err != nil {
		return fmt.Errorf("generating values.yaml failed: %v", err)
	}
	if err := generateChartYaml(version); err != nil {
		return fmt.Errorf("generating Chart.yaml failed: %v", err)
	}
	return nil
}

func getOsGlooVersion(prefix string) (string, error) {
	tomlTree, err := parseToml(prefix)
	if err != nil {
		return "", err
	}
	version, err := versionutils.GetVersion(versionutils.GlooPkg, tomlTree)
	if err != nil {
		return "", fmt.Errorf("failed to determine open source Gloo version. Cause: %v", err)
	}
	log.Printf("Open source gloo version is: %v", version)
	return version, nil
}

func parseToml(prefix string) ([]*toml.Tree, error) {
	tomlPath := filepath.Join(prefix, gopkgToml)
	config, err := toml.LoadFile(tomlPath)
	if err != nil {
		return nil, err
	}

	tomlTree := config.Get(constraint)

	switch typedTree := tomlTree.(type) {
	case []*toml.Tree:
		return typedTree, nil
	default:
		return nil, fmt.Errorf("unable to parse toml tree")
	}
}

func readConfig(path string) (Config, error) {
	var config Config
	if err := readYaml(path, &config); err != nil {
		return config, err
	}
	return config, nil
}

func readYaml(path string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.Unmarshal(bytes, obj); err != nil {
		return errors.Wrap(err, "failed parsing configuration file")
	}

	return nil
}

func writeYaml(obj interface{}, path string) error {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "failed marshaling config struct")
	}

	err = ioutil.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failing writing config file")
	}
	return nil
}

func generateValuesYaml(imageTag, imageRepoPrefix, glooVersion string) error {
	config, err := readConfig(filepath.Join(rootPrefix, DefaultValues))
	if err != nil {
		return err
	}

	config.Supergloo.Deployment.Image.Repository = path.Join(imageRepoPrefix, constants.SuperglooImageName)
	config.Supergloo.Deployment.Image.Tag = imageTag

	config.SidecarInjector.Image.Repository = path.Join(imageRepoPrefix, constants.SidecarInjectorImageName)
	config.SidecarInjector.Image.Tag = imageTag

	config.MeshDiscovery.Deployment.Image.Repository = path.Join(imageRepoPrefix, constants.MeshDiscoveryImageName)
	config.MeshDiscovery.Deployment.Image.Tag = imageTag

	config.Discovery.Deployment.Image.Tag = glooVersion

	return writeYaml(&config, filepath.Join(rootPrefix, ValuesOutput))
}

func generateChartYaml(version string) error {
	var chart glooGenerate.Chart
	if err := readYaml(filepath.Join(rootPrefix, ChartTemplate), &chart); err != nil {
		return err
	}

	chart.Version = version

	return writeYaml(&chart, filepath.Join(rootPrefix, ChartOutput))
}
