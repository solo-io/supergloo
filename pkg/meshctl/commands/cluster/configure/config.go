package configure

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

const managementPlane = "managementPlane"

type Config struct {
	ApiVersion string                 `json:"apiVersion"`
	Clusters   map[string]KubeCluster `json:"clusters"`
}

type KubeCluster struct {
	KubeConfig  string `json:"kubeConfig"`
	KubeContext string `json:"kubeContext"`
}

func ParseMeshctlConfig(fileName string) (Config, error) {
	config := Config{}
	_, fileErr := os.Stat(fileName)
	if os.IsExist(fileErr) || fileErr == nil {
		contentString, err := ioutil.ReadFile(fileName)
		if err != nil {
			return config, err
		}
		if err := yaml.Unmarshal(contentString, &config); err != nil {
			return config, err
		}
	}
	return config, nil
}

func MeshctlConfigFilePath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, ".gloo-mesh", "meshctl-config.yaml"), nil
}
