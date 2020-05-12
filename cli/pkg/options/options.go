package options

import (
	"time"
)

// wire provider func, not meant to be used outside of that
func NewOptionsProvider() *Options {
	return &Options{}
}

type Options struct {
	Root         Root
	Cluster      Cluster
	Upgrade      Upgrade
	SmhInstall   SmhInstall
	Mesh         Mesh
	Check        Check
	Get          Get
	SmhUninstall SmhUninstall
	Demo         Demo
	Describe     Describe
	Create       Create
}

type Root struct {
	KubeConfig     string
	KubeContext    string
	WriteNamespace string
	KubeTimeout    time.Duration
	Verbose        bool
}

type Cluster struct {
	Register   Register
	Deregister Deregister
}

type Register struct {
	RemoteClusterName          string
	RemoteWriteNamespace       string
	RemoteContext              string
	RemoteKubeConfig           string
	LocalClusterDomainOverride string
	Overwrite                  bool
	UseDevCsrAgentChart        bool
}

type Deregister struct {
	RemoteClusterName string
}

type Mesh struct {
	Install MeshInstall
}

type MeshInstall struct {
	InstallationConfig MeshInstallationConfig
	DryRun             bool
	ManifestPath       string
	Profile            string
}

type MeshInstallationConfig struct {
	CreateNamespace  bool
	InstallNamespace string
}

type Upgrade struct {
	ReleaseTag   string
	DownloadPath string
}

type SmhInstall struct {
	DryRun                  bool
	HelmChartOverride       string
	HelmChartValueFileNames []string
	HelmReleaseName         string
	Version                 string
	CreateNamespace         bool
	Register                bool
	ClusterName             string
}

type SmhUninstall struct {
	ReleaseName     string
	RemoveNamespace bool
}

type Check struct {
	OutputFormat string
}

type Get struct {
	OutputFormat string
}

type Describe struct {
	Policies string
}

type Demo struct {
	DemoLabel         string
	UseKind           bool
	ClusterName       string
	IstioMulticluster IstioMulticluster
	DevMode           bool
	ContextName       string
}

type IstioMulticluster struct {
	RemoteClusterName string
	RemoteContextName string
}

type Create struct {
	Interactive  bool
	DryRun       bool
	ResourceType string
	VirtualMesh  CreateVirtualMesh
	OutputFormat string
}

type CreateVirtualMesh struct {
	ForAllMeshes bool
	Meshes       []string
}
