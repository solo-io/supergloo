package operator

import k8s_apps_v1 "k8s.io/api/apps/v1"

type IstioVersion string

const (
	IstioVersion1_5 IstioVersion = "1.5.1"
	IstioVersion1_6 IstioVersion = "1.6.0"

	DefaultIstioProfile = "default"

	DefaultIstioOperatorNamespace      = "istio-operator"
	DefaultIstioOperatorDeploymentName = "istio-operator"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// install and manage an instance of the Istio operator
type OperatorManager interface {
	// we try to make the install attempt atomic; if one of the resources fails to install, the previous successful ones are cleaned up
	InstallOperatorApplication(installationOptions *InstallationOptions) error

	// produce a string representation of the application manifest (not including operator config) that would be applied by calling InstallOperatorApplication
	InstallDryRun(installationOptions *InstallationOptions) (manifest string, err error)

	// get the yaml manifest for the oeprator config that would be applied
	OperatorConfigDryRun(installationOptions *InstallationOptions) (manifest string, err error)
}

type InstallationOptions struct {
	IstioVersion IstioVersion

	// the default defined above is used here if no value provided
	InstallNamespace string

	CreateNamespace bool

	// optional; the name of one of Istio's pre-set installation profile to use
	// if neither this field nor CustomIstioOperatorSpec are provided, "default" is used here
	// takes precedence over CustomIstioOperatorSpec if both are provided
	InstallationProfile string

	// optional; a yaml representation of a customized operator spec
	CustomIstioOperatorSpec string
}

type InstallerManifestBuilder interface {
	// Based on the pending installation config, generate an appropriate installation manifest
	BuildOperatorDeploymentManifest(istioVersion IstioVersion, installNamespace string, createNamespace bool) (string, error)

	// Generate an IstioOperator spec that sets up Mesh with its demo profile
	BuildOperatorConfigurationWithProfile(profile, installationNamespace string) (string, error)
}

type OperatorDao interface {
	// applies a YAML string representation of the operator manifest
	ApplyManifest(installationNamespace, manifest string) error

	// returns (nil, nil) if not found
	FindOperatorDeployment(name, namespace string) (*k8s_apps_v1.Deployment, error)
}
