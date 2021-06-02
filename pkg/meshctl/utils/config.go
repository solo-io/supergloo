package utils

import (
	"fmt"
	"github.com/rotisserie/eris"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

const managementPlane = "managementPlane"

type Config struct {
	filepath   string
	ApiVersion string                 `json:"apiVersion"`
	Clusters   map[string]KubeCluster `json:"clusters"`
}

// returns the path of the file storing the config
func (c Config) FilePath() string {
	return c.filepath
}

// returns the path of the file storing the config
func (c Config) MgmtCluster() KubeCluster {
	return c.Clusters[managementPlane]
}

// returns the path of the file storing the config
func (c Config) AddMgmtCluster(kc KubeCluster) {
	c.Clusters[managementPlane] = kc
}

// returns the path of the file storing the config
func (c Config) AddDataPlaneCluster(name string, kc KubeCluster) error {
	if name == managementPlane {
		return eris.Errorf("%v is a special cluster name reserved for the management cluster. try a different name", name)
	}
	c.Clusters[name] = kc
	return nil
}

type KubeCluster struct {
	KubeConfig  string `json:"kubeConfig"`
	KubeContext string `json:"kubeContext"`
}

func ParseMeshctlConfig(meshctlConfigPath string) (Config, error) {
	if meshctlConfigPath == "" {
		var err error
		meshctlConfigPath, err = meshctlConfigFilePath()
		if err != nil {
			return Config{}, err
		}
	}

	config := Config{}

	if _, fileErr := os.Stat(meshctlConfigPath); fileErr == nil {
		contentString, err := ioutil.ReadFile(meshctlConfigPath)
		if err != nil {
			return config, err
		}
		if err := yaml.Unmarshal(contentString, &config); err != nil {
			return config, err
		}
	}
	config.filepath = meshctlConfigPath
	if config.ApiVersion == "" {
		config.ApiVersion = "v1"
	}
	if config.Clusters == nil {
		config.Clusters = map[string]KubeCluster{}
	}

	if config.ApiVersion != "v1" {
		return Config{}, fmt.Errorf("unrecognized api version: %v", config.ApiVersion)
	}


	return config, nil
}

func meshctlConfigFilePath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, ".gloo-mesh", "meshctl-config.yaml"), nil
}
