package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/solo-io/go-utils/versionutils"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	glooGenerate "github.com/solo-io/gloo/install/helm/gloo/generate"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/supergloo/install/helm/supergloo/generate"
)

const (
	defaultValues = "install/helm/supergloo/values-defaults.yaml"
	valuesOutput  = "install/helm/supergloo/values.yaml"
	chartTemplate = "install/helm/supergloo/Chart-template.yaml"
	chartOutput   = "install/helm/supergloo/Chart.yaml"

	glooPkg = "github.com/solo-io/gloo"

	neverPull    = "Never"
	alwaysPull   = "Always"
	ifNotPresent = "IfNotPresent"
)

var (
	osGlooVersion string
)

func main() {
	var version string
	if len(os.Args) < 2 {
		panic("Must provide version as argument")
	} else {
		version = os.Args[1]
	}
	log.Printf("Generating helm files.")
	pullPolicy := alwaysPull
	if version == "dev" {
		pullPolicy = neverPull
	}

	if err := getOsGlooVersion(); err != nil {
		panic(err)
	}

	if err := generateValuesYaml(version, pullPolicy, valuesOutput); err != nil {
		log.Fatalf("generating values.yaml failed: %v", err)
	}
	if err := generateChartYaml(version); err != nil {
		log.Fatalf("generating Chart.yaml failed: %v", err)
	}
}

func getOsGlooVersion() error {
	tomlTree, err := versionutils.ParseToml()
	if err != nil {
		panic(err)
	}
	version, err := versionutils.GetVersion(glooPkg, tomlTree)
	if err != nil {
		return fmt.Errorf("failed to determine open source Gloo version. Cause: %v", err)
	}
	log.Printf("Open source gloo version is: %v", version)
	osGlooVersion = version
	return nil
}

func readConfig(path string) (generate.Config, error) {
	var config generate.Config
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

func generateValuesYaml(version, pullPolicy, outputFile string) error {
	config, err := readConfig(defaultValues)
	if err != nil {
		return err
	}
	config.Supergloo.Deployment.Image.Tag = version
	config.Supergloo.Deployment.Image.PullPolicy = pullPolicy

	config.Discovery.Deployment.Image.Tag = osGlooVersion
	config.Discovery.Deployment.Image.PullPolicy = pullPolicy

	return writeYaml(&config, outputFile)
}

func generateChartYaml(version string) error {
	var chart glooGenerate.Chart
	if err := readYaml(chartTemplate, &chart); err != nil {
		return err
	}

	chart.Version = version

	return writeYaml(&chart, chartOutput)
}
