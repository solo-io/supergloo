package generate

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/solo-io/go-utils/versionutils"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	glooGenerate "github.com/solo-io/gloo/install/helm/gloo/generate"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

const (
	DefaultValues = "install/helm/supergloo/values-defaults.yaml"
	ValuesOutput  = "install/helm/supergloo/values.yaml"
	ChartTemplate = "install/helm/supergloo/Chart-template.yaml"
	ChartOutput   = "install/helm/supergloo/Chart.yaml"
)

var (
	osGlooVersion string
)

func Run(version, pullPolicy string) error {
	if glooVersion, err := getOsGlooVersion(); err != nil {
		return err
	} else {
		osGlooVersion = glooVersion
	}

	if err := generateValuesYaml(version, pullPolicy); err != nil {
		return fmt.Errorf("generating values.yaml failed: %v", err)
	}
	if err := generateChartYaml(version); err != nil {
		return fmt.Errorf("generating Chart.yaml failed: %v", err)
	}
	return nil
}

func getOsGlooVersion() (string, error) {
	tomlTree, err := versionutils.ParseToml()
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

func generateValuesYaml(version, pullPolicy string) error {
	config, err := readConfig(DefaultValues)
	if err != nil {
		return err
	}
	config.Supergloo.Deployment.Image.Tag = version
	config.Supergloo.Deployment.Image.PullPolicy = pullPolicy

	config.Discovery.Deployment.Image.Tag = osGlooVersion
	config.Discovery.Deployment.Image.PullPolicy = pullPolicy

	return writeYaml(&config, ValuesOutput)
}

func generateChartYaml(version string) error {
	var chart glooGenerate.Chart
	if err := readYaml(ChartTemplate, &chart); err != nil {
		return err
	}

	chart.Version = version

	return writeYaml(&chart, ChartOutput)
}
