package operator

import (
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common/kube"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator/install"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	FailedToParseInstallManifest = func(err error) error {
		return eris.Wrap(err, "Failed to parse the templated Istio operator manifest.")
	}
	FailedToInstallOperator = func(err error) error {
		return eris.Wrap(err, "Failed to install the Istio installation operator. Artifacts from the failed installation will be cleaned up.")
	}
	FailedToCleanFailedInstallation = func(err error) error {
		return eris.Wrap(err, "Failed to clean up the failed Istio operator installation.")
	}
	OperatorVersionMismatch = func(clusterName, installNamespace, specifiedVersion, actualVersion string) error {
		return eris.Errorf(
			"Istio operator is already running in cluster '%s' namespace '%s', but its version does not match what was specified; specified '%s' but found '%s'",
			clusterName,
			installNamespace,
			specifiedVersion,
			actualVersion,
		)
	}
	UnrecognizedOperatorInstance    = eris.New("This instance of the Istio operator is not recognized - cannot verify its version matches what you specified")
	FailedToGenerateInstallManifest = func(err error) error {
		return eris.Wrap(err, "Install manifest template failed to render. This shouldn't happen")
	}
	FailedToValidateExistingOperator = func(err error, clusterName, namespace, version string) error {
		return eris.Wrapf(err, "Failed to validate that the existing Istio operator deployment in cluster %s namespace %s is the requested version: %s", clusterName, namespace, version)
	}
	FailedToCheckIfOperatorExists = func(err error, clusterName, namespace, version string) error {
		return eris.Wrapf(err, "Failed to check whether the Istio operator is already installed to cluster %s in namespace %s at version %s", clusterName, namespace, version)
	}

	DefaultIstioOperatorNamespace      = "istio-operator"
	DefaultIstioOperatorVersion        = "1.5.0-alpha.0"
	DefaultIstioOperatorDeploymentName = "istio-operator"
)

//go:generate mockgen -source manager.go -destination ./mocks/mock_operator_manager.go
type OperatorManager interface {
	// install an instance of the Istio operator
	// we try to make the install attempt atomic; if one of the resources fails to install, the previous successful ones are cleaned up
	Install() error

	// ensure that the given namespace is appropriate for installing an Istio operator
	// will fail with an error if the operator is already present at a different version than we're specifying
	// will return (true, nil) if no operator deployment is present yet
	// (false, nil) indicates the operator is present already at an appropriate version, so no need to call .Install()
	//
	// the `clusterName` arg is only used for error reporting
	ValidateOperatorNamespace(clusterName string) (installNeeded bool, err error)
}

type OperatorManagerFactory func(
	unstructuredKubeClient kube.UnstructuredKubeClient,
	manifestBuilder InstallerManifestBuilder,
	deploymentClient server.DeploymentClient,
	imageNameParser docker.ImageNameParser,
	installationConfig *install.InstallationConfig,
) OperatorManager

func NewOperatorManagerFactory() OperatorManagerFactory {
	return NewManager
}

func NewManager(
	unstructuredKubeClient kube.UnstructuredKubeClient,
	manifestBuilder InstallerManifestBuilder,
	deploymentClient server.DeploymentClient,
	imageNameParser docker.ImageNameParser,
	installationConfig *install.InstallationConfig,
) OperatorManager {

	setDefaults(installationConfig)

	return &manager{
		unstructuredKubeClient: unstructuredKubeClient,
		manifestBuilder:        manifestBuilder,
		deploymentClient:       deploymentClient,
		imageNameParser:        imageNameParser,
		installationConfig:     installationConfig,
	}
}

type manager struct {
	unstructuredKubeClient kube.UnstructuredKubeClient
	manifestBuilder        InstallerManifestBuilder
	deploymentClient       server.DeploymentClient
	imageNameParser        docker.ImageNameParser
	installationConfig     *install.InstallationConfig
}

func (m *manager) Install() error {
	installationManifest, err := m.manifestBuilder.Build(m.installationConfig)
	if err != nil {
		return FailedToGenerateInstallManifest(err)
	}

	resources, err := m.unstructuredKubeClient.BuildResources(m.installationConfig.InstallNamespace, installationManifest)
	if err != nil {
		return FailedToParseInstallManifest(err)
	}

	createdResources, installErr := m.unstructuredKubeClient.Create(m.installationConfig.InstallNamespace, resources)
	if installErr != nil {
		_, deleteErr := m.unstructuredKubeClient.Delete(m.installationConfig.InstallNamespace, createdResources)
		if deleteErr != nil {
			var multiErr *multierror.Error
			multiErr = multierror.Append(multiErr, FailedToInstallOperator(installErr))
			multiErr = multierror.Append(multiErr, FailedToCleanFailedInstallation(deleteErr))

			return multiErr
		}

		return FailedToInstallOperator(installErr)
	}

	return nil
}

func (m *manager) ValidateOperatorNamespace(clusterName string) (installNeeded bool, err error) {
	deployments, err := m.deploymentClient.GetDeployments(m.installationConfig.InstallNamespace, "")
	if err != nil {
		return false, FailedToCheckIfOperatorExists(err, clusterName, m.installationConfig.InstallNamespace, m.installationConfig.IstioOperatorVersion)
	}

	if deployments == nil {
		return true, nil
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == DefaultIstioOperatorDeploymentName {
			err := m.validateExistingOperatorDeployment(clusterName, m.installationConfig, deployment)
			if err != nil {
				return false, FailedToValidateExistingOperator(err, clusterName, m.installationConfig.InstallNamespace, m.installationConfig.IstioOperatorVersion)
			}

			// no install needed, and no error occurred
			return false, nil
		}
	}

	// no deployment was found, and no error occurred
	return true, nil
}

func (m *manager) validateExistingOperatorDeployment(clusterName string, installerOptions *install.InstallationConfig, deployment appsv1.Deployment) error {
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return UnrecognizedOperatorInstance
	}

	image, err := m.imageNameParser.Parse(containers[0].Image)
	if err != nil {
		return err
	}

	actualImageVersion := image.Tag
	if actualImageVersion == "" {
		actualImageVersion = image.Digest
	}

	if actualImageVersion != installerOptions.IstioOperatorVersion {
		return OperatorVersionMismatch(clusterName, installerOptions.InstallNamespace, installerOptions.IstioOperatorVersion, actualImageVersion)
	}

	return nil
}

func setDefaults(installerOptions *install.InstallationConfig) {
	if installerOptions.InstallNamespace == "" {
		installerOptions.InstallNamespace = DefaultIstioOperatorNamespace
	}

	if installerOptions.IstioOperatorVersion == "" {
		installerOptions.IstioOperatorVersion = DefaultIstioOperatorVersion
	}
}
