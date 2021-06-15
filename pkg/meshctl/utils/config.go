package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rotisserie/eris"

	"github.com/ghodss/yaml"
)

const managementPlane = "managementPlane"

type MeshctlConfig struct {
	filepath   string
	ApiVersion string                    `json:"apiVersion"`
	Clusters   map[string]MeshctlCluster `json:"clusters"`
}

type MeshctlCluster struct {
	KubeConfig  string `json:"kubeConfig"`
	KubeContext string `json:"kubeContext"`
}

type KubeConfig struct {
	ApiVersion string              `json:"apiVersion"`
	Clusters   []KubeConfigCluster `yaml:"clusters"`
}

type KubeConfigCluster struct {
	Name string `yaml:"name"`
}

// returns the path of the file storing the config
func (c MeshctlConfig) FilePath() string {
	return c.filepath
}

// returns the path of the file storing the config
func (c MeshctlConfig) MgmtCluster() MeshctlCluster {
	return c.Clusters[managementPlane]
}

// returns the path of the file storing the config
func (c MeshctlConfig) AddMgmtCluster(kc MeshctlCluster) {
	c.Clusters[managementPlane] = kc
}

// returns the path of the file storing the config
func (c MeshctlConfig) AddDataPlaneCluster(name string, kc MeshctlCluster) error {
	if name == managementPlane {
		return eris.Errorf("%v is a special cluster name reserved for the management cluster. try a different name", name)
	}
	c.Clusters[name] = kc
	return nil
}

func ParseMeshctlConfig(meshctlConfigPath string) (MeshctlConfig, error) {
	if meshctlConfigPath == "" {
		var err error
		meshctlConfigPath, err = meshctlConfigFilePath()
		if err != nil {
			return MeshctlConfig{}, err
		}
	}

	config := MeshctlConfig{}

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
		config.Clusters = map[string]MeshctlCluster{}
	}

	if config.ApiVersion != "v1" {
		return MeshctlConfig{}, fmt.Errorf("unrecognized api version: %v", config.ApiVersion)
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
