package options

import (
	"time"
)

// wire provider func, not meant to be used outside of that
func NewOptionsProvider() *Options {
	return &Options{}
}

type Options struct {
	Root       Root
	Cluster    Cluster
	Upgrade    Upgrade
	SmhInstall SmhInstall
	Istio      Istio
}

type Root struct {
	KubeConfig     string
	KubeContext    string
	WriteNamespace string
	KubeTimeout    time.Duration
	Verbose        bool
}

type Cluster struct {
	Register Register
}

type Register struct {
	RemoteClusterName          string
	RemoteWriteNamespace       string
	RemoteContext              string
	RemoteKubeConfig           string
	LocalClusterDomainOverride string
}

type Istio struct {
	Install IstioInstall
}

type IstioInstall struct {
	InstallationConfig            IstioInstallationConfig
	DryRun                        bool
	IstioControlPlaneManifestPath string
	Profile                       string
}

type IstioInstallationConfig struct {
	CreateNamespace            bool
	CreateIstioControlPlaneCRD bool

	// will be defaulted to istio-operator if left blank
	InstallNamespace string

	// will be defaulted to `DefaultIstioOperatorVersion` if left blank
	IstioOperatorVersion string
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
}
