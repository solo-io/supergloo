package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	gloo_generate "github.com/solo-io/gloo/install/helm/gloo/generate"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/mesh-projects/install/helm/mesh-projects/generate"
)

var (
	valuesTemplate = "install/helm/mesh-projects/values-template.yaml"
	valuesOutput   = "install/helm/mesh-projects/values.yaml"
	docsOutput     = "docs/helm-values.md"
	chartTemplate  = "install/helm/mesh-projects/Chart-template.yaml"
	chartOutput    = "install/helm/mesh-projects/Chart.yaml"

	always     = "Always"
	constraint = "constraint"

	rootPrefix = ""
	gopkgToml  = "Gopkg.toml"
	glooPkg    = "github.com/solo-io/gloo"
)

func main() {
	var version, repoPrefixOverride, globalPullPolicy string
	if len(os.Args) < 2 {
		panic("Must provide version as argument")
	} else {
		version = os.Args[1]

		if len(os.Args) >= 3 {
			repoPrefixOverride = os.Args[2]
		}
		if len(os.Args) >= 4 {
			globalPullPolicy = os.Args[3]
		}
	}

	glooVersion, err := getOsGlooVersion(rootPrefix)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("Generating helm files.")
	if err := generateValuesYaml(version, repoPrefixOverride, globalPullPolicy, glooVersion); err != nil {
		log.Fatalf("generating values.yaml failed!: %v", err)
	}
	if err := generateChartYaml(version); err != nil {
		log.Fatalf("generating Chart.yaml failed!: %v", err)
	}
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

func writeDocs(docs helmchart.HelmValues, path string) error {
	err := ioutil.WriteFile(path, []byte(docs.ToMarkdown()), os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failing writing helm values file")
	}
	return nil
}

func readConfig() (*generate.HelmConfig, error) {
	var config generate.HelmConfig
	if err := readYaml(valuesTemplate, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func generateValuesYaml(version, repositoryPrefix, globalPullPolicy, glooVersion string) error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}

	cfg.MeshBridge.Deployment.Image.Tag = version
	cfg.MeshDiscovery.Deployment.Image.Tag = version
	cfg.Discovery.Deployment.Image.Tag = glooVersion

	if version == "dev" {
		cfg.MeshBridge.Deployment.Image.PullPolicy = always
		cfg.MeshDiscovery.Deployment.Image.PullPolicy = always
	}

	if repositoryPrefix != "" {
		cfg.Global.Image.Registry = repositoryPrefix
	}

	if globalPullPolicy != "" {
		cfg.Global.Image.PullPolicy = globalPullPolicy
	}

	if err := writeDocs(helmchart.Doc(cfg), docsOutput); err != nil {
		return err
	}

	return writeYaml(cfg, valuesOutput)
}

func generateChartYaml(version string) error {
	var chart gloo_generate.Chart
	if err := readYaml(chartTemplate, &chart); err != nil {
		return err
	}

	chart.Version = version

	return writeYaml(&chart, chartOutput)
}

func getOsGlooVersion(prefix string) (string, error) {
	tomlTree, err := parseToml(prefix)
	if err != nil {
		return "", err
	}
	version, err := versionutils.GetVersion(glooPkg, tomlTree)
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
