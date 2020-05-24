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

	// produce a string representation of the total manifest (application plus operator config) that would be applied by calling InstallOperatorApplication
	InstallDryRun(installationOptions *InstallationOptions) (manifest string, err error)

	// ensure that the given namespace is appropriate for installing an Mesh operator
	// will fail with an error if the operator is already present at a different version than we're specifying
	// will return (true, nil) if no operator deployment is present yet
	// (false, nil) indicates the operator is present already at an appropriate version, so no need to call .Install()
	//
	// the `clusterName` arg is only used for error reporting
	ValidateOperatorNamespace(istioVersion IstioVersion, operatorName, operatorNamespace, clusterName string) (installNeeded bool, err error)
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
