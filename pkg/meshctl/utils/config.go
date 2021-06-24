package utils

import (
	"os"
	"path"

	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
)

const (
	ConfigFileName = "meshctl-config.yaml"
	ConfigDirName  = ".gloo-mesh"

	// this is a workaround - we can't set cobra's default arg to "$HOME/..." and have it just work, because
	// it doesn't expand $HOME. We also can't set the default value to the expanded value of $HOME, ie something like
	// os.UserHomeDir(), because that will change the content of our generated docs/ directory based on whatever system
	// built glooctl last. So we settle for this placeholder.
	homeDir = "<home_directory>"

	managementPlane = "managementPlane"
)

var (
	DefaultConfigPath           = path.Join(homeDir, ConfigDirName, ConfigFileName)
	invalidMeshctlConfigFileErr = eris.New("please either configure or pass in a valid meshctl config file (see the 'meshctl cluster config' command)")
)

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
	meshctlConfigFile, err := DefaultMeshctlConfigFilePath()
	if err != nil {
		return err
	}
	// if the existing meschtl config file matches, update it
	config, err := ParseMeshctlConfig(meshctlConfigFile)
	if err != nil {
		config = NewMeshctlConfig()
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
	return WriteConfigToFile(config, "")
}

// update the default meschtl config file with deregistration info
func UpdateMeshctlConfigWithDeregistrationInfo(mgmtKubeConfigPath, mgmtKubecontext,
	remoteClusterName, remoteKubeConfigPath string) error {
	meshctlConfigFile, err := DefaultMeshctlConfigFilePath()
	if err != nil {
		return err
	}
	// if the existing meschtl config file doesn't match, don't modify it
	config, err := ParseMeshctlConfig(meshctlConfigFile)
	if err != nil {
		config = NewMeshctlConfig()
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
	return WriteConfigToFile(config, "")
}

// update the default meschtl config file with install info
func UpdateMeshctlConfigWithInstallInfo(mgmtKubeConfig, mgmtKubecontext string) error {
	meshctlConfigFile, err := DefaultMeshctlConfigFilePath()
	if err != nil {
		return err
	}
	// Start over with a new config
	config := NewMeshctlConfig()
	config.AddMgmtCluster(MeshctlCluster{
		KubeConfig:  mgmtKubeConfig,
		KubeContext: mgmtKubecontext,
	})
	return WriteConfigToFile(config, meshctlConfigFile)
}

// parse the meshctl config file into a MeshctlConfig struct
// If the file doesn't exist or is invalid, return an error
func ParseMeshctlConfig(meshctlConfigPath string) (MeshctlConfig, error) {
	config := NewMeshctlConfig()
	if meshctlConfigPath == DefaultConfigPath {
		var err error
		meshctlConfigPath, err = DefaultMeshctlConfigFilePath()
		if err != nil {
			return config, err
		}
	}
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
	if config.ApiVersion != "v1" {
		return config, eris.Errorf("meshctl config file has an unrecognized api version: %v", config.ApiVersion)
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
	if meshctlConfigPath == DefaultConfigPath {
		var err error
		meshctlConfigPath, err = DefaultMeshctlConfigFilePath()
		if err != nil {
			return err
		}
	}
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
