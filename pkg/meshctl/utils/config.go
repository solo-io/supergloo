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

var invalidMeshctlConfigFileErr = eris.New("please either configure or pass in a valid meshctl config file (see the 'meshctl cluster config' command)")

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

// returns the mgmt meshctl cluster config
func (c MeshctlConfig) MgmtCluster() MeshctlCluster {
	if mgmtCluster, ok := c.Clusters[managementPlane]; ok {
		return mgmtCluster
	}
	return MeshctlCluster{}
}

// returns the mgmt meshctl cluster name
func (c MeshctlConfig) IsMgmtCluster(name string) bool {
	return name == managementPlane
}

// add the management cluster config
func (c MeshctlConfig) AddMgmtCluster(kc MeshctlCluster) {
	c.Clusters[managementPlane] = kc
}

// add a data plane cluster config
func (c MeshctlConfig) AddDataPlaneCluster(name string, kc MeshctlCluster) error {
	if name == managementPlane {
		return eris.Errorf("%v is a special cluster name reserved for the management cluster. try a different name", name)
	}
	c.Clusters[name] = kc
	return nil
}

// update the default meschtl config file with registration info
func UpdateMeshctlConfigWithRegistrationInfo(mgmtKubeConfigPath, mgmtKubecontext,
	remoteClusterName, remoteKubeConfigPath, remoteKubecontext string) error {
	// if the existing meschtl config file matches, update it
	config := NewMeshctlConfig()
	meshctlConfigFile, err := DefaultMeshctlConfigFilePath()
	if err == nil {
		if parsedConfig, err := ParseMeshctlConfig(meshctlConfigFile); err == nil {
			config = parsedConfig
		}
	}
	kubeConfig := mgmtKubeConfigPath
	if kubeConfig == "" {
		kubeConfig = remoteKubeConfigPath
	}
	if config.MgmtCluster().KubeConfig != kubeConfig || config.MgmtCluster().KubeContext != mgmtKubecontext {
		// Otherwise, start over with a new config
		config = NewMeshctlConfig()
		config.AddMgmtCluster(MeshctlCluster{
			KubeConfig:  kubeConfig,
			KubeContext: mgmtKubecontext,
		})
	}
	err = config.AddDataPlaneCluster(remoteClusterName, MeshctlCluster{
		KubeConfig:  remoteKubeConfigPath,
		KubeContext: remoteKubecontext,
	})
	if err != nil {
		return err
	}
	return WriteConfigToFile(config, meshctlConfigFile)
}

// update the default meschtl config file with deregistration info
func UpdateMeshctlConfigWithDeregistrationInfo(mgmtKubeConfigPath, mgmtKubecontext,
	remoteClusterName, remoteKubeConfigPath string) error {
	// if the existing meschtl config file doesn't match, don't modify it
	config := NewMeshctlConfig()
	meshctlConfigFile, err := DefaultMeshctlConfigFilePath()
	if err == nil {
		if parsedConfig, err := ParseMeshctlConfig(meshctlConfigFile); err == nil {
			config = parsedConfig
		}
	}
	kubeConfig := mgmtKubeConfigPath
	if kubeConfig == "" {
		kubeConfig = remoteKubeConfigPath
	}
	if config.MgmtCluster().KubeConfig != kubeConfig || config.MgmtCluster().KubeContext != mgmtKubecontext {
		return nil
	}
	// Otherwise, update it
	delete(config.Clusters, remoteClusterName)
	return WriteConfigToFile(config, meshctlConfigFile)
}

// update the default meschtl config file with install info
func UpdateMeshctlConfigWithInstallInfo(mgmtKubeConfig, mgmtKubecontext string) error {
	meshctlConfigFile, err := DefaultMeshctlConfigFilePath()
	if err != nil {
		return err
	}
	// Start over with a new config
	config := MeshctlConfig{}
	config.AddMgmtCluster(MeshctlCluster{
		KubeConfig:  mgmtKubeConfig,
		KubeContext: mgmtKubecontext,
	})
	return WriteConfigToFile(config, meshctlConfigFile)
}

// parse the meshctl config file into a MeshctlConfig struct
func ParseMeshctlConfig(meshctlConfigPath string) (MeshctlConfig, error) {
	config := NewMeshctlConfig()
	if _, err := os.Stat(meshctlConfigPath); err != nil {
		return config, invalidMeshctlConfigFileErr
	}
	contentString, err := ioutil.ReadFile(meshctlConfigPath)
	if err != nil {
		return config, invalidMeshctlConfigFileErr
	}
	if err := yaml.Unmarshal(contentString, &config); err != nil {
		return config, invalidMeshctlConfigFileErr
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

// new initialized meshctl config
func NewMeshctlConfig() MeshctlConfig {
	return MeshctlConfig{
		ApiVersion: "v1",
		Clusters:   map[string]MeshctlCluster{managementPlane: MeshctlCluster{KubeConfig: "", KubeContext: ""}},
	}
}

func WriteConfigToFile(config MeshctlConfig, meshctlConfigPath string) error {
	bytes, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(meshctlConfigPath, bytes, 0644); err != nil {
		return err
	}
	return err
}

// return the default meshctl config filepath
func DefaultMeshctlConfigFilePath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, ".gloo-mesh", "meshctl-config.yaml"), nil
}
