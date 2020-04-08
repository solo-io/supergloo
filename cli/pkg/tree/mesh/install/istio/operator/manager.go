package operator

import (
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	FailedToParseInstallManifest = func(err error) error {
		return eris.Wrap(err, "Failed to parse the templated Mesh operator manifest.")
	}
	FailedToInstallOperator = func(err error) error {
		return eris.Wrap(err, "Failed to install the Mesh installation operator. Artifacts from the failed installation will be cleaned up.")
	}
	FailedToCleanFailedInstallation = func(err error) error {
		return eris.Wrap(err, "Failed to clean up the failed Mesh operator installation.")
	}
	UnrecognizedOperatorInstance    = eris.New("This instance of the Mesh operator is not recognized - cannot verify its version matches what you specified")
	FailedToGenerateInstallManifest = func(err error) error {
		return eris.Wrap(err, "Install manifest template failed to render. This shouldn't happen")
	}
	FailedToValidateExistingOperator = func(err error, clusterName, namespace string) error {
		return eris.Wrapf(err, "Failed to validate existing Mesh operator deployment in cluster %s namespace %s", clusterName, namespace)
	}
	FailedToDetermineOperatorVersion = func(current string) error {
		return eris.Errorf("Failed to determine whether the current operator running at version %s is the minimum version %s", current, MinimumOperatorVersion.String())
	}
	IncompatibleOperatorVersion = func(current string) error {
		return eris.Errorf("Found istio operator running at version %s, while %s is the minimum supported version", current, MinimumOperatorVersion.String())
	}
	FailedToCheckIfOperatorExists = func(err error, clusterName, namespace string) error {
		return eris.Wrapf(err, "Failed to check whether the Mesh operator is already installed to cluster %s in namespace %s", clusterName, namespace)
	}

	MinimumOperatorVersion = versionutils.Version{
		Major: 1,
		Minor: 5,
		Patch: 1,
	}
)

//go:generate mockgen -source manager.go -destination ./mocks/mock_operator_manager.go
type OperatorManager interface {
	// install an instance of the Mesh operator
	// we try to make the install attempt atomic; if one of the resources fails to install, the previous successful ones are cleaned up
	Install() error

	// ensure that the given namespace is appropriate for installing an Mesh operator
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
	installationConfig *options.MeshInstallationConfig,
) OperatorManager

func NewOperatorManagerFactory() OperatorManagerFactory {
	return NewManager
}

func NewManager(
	unstructuredKubeClient kube.UnstructuredKubeClient,
	manifestBuilder InstallerManifestBuilder,
	deploymentClient server.DeploymentClient,
	imageNameParser docker.ImageNameParser,
	installationConfig *options.MeshInstallationConfig,
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
	installationConfig     *options.MeshInstallationConfig
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
		return false, FailedToCheckIfOperatorExists(err, clusterName, m.installationConfig.InstallNamespace)
	}

	if deployments == nil {
		return true, nil
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == cliconstants.DefaultIstioOperatorDeploymentName {
			if err = m.validateExistingOperatorDeployment(clusterName, m.installationConfig, deployment); err != nil {
				return false, FailedToValidateExistingOperator(err, clusterName, m.installationConfig.InstallNamespace)
			}

			// no install needed, and no error occurred
			return false, nil
		}
	}

	// no deployment was found, and no error occurred
	return true, nil
}

func (m *manager) validateExistingOperatorDeployment(clusterName string, installerOptions *options.MeshInstallationConfig, deployment appsv1.Deployment) error {
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

	if !strings.HasPrefix(actualImageVersion, "v") {
		actualImageVersion = "v" + actualImageVersion
	}

	version, err := versionutils.ParseVersion(actualImageVersion)
	if err != nil {
		return err
	}

	greater, ok := version.IsGreaterThan(MinimumOperatorVersion)
	if !ok {
		return FailedToDetermineOperatorVersion(version.String())
	}

	if !(greater || version.Equals(&MinimumOperatorVersion)) {
		return IncompatibleOperatorVersion(version.String())
	}

	return nil
}

func setDefaults(installerOptions *options.MeshInstallationConfig) {
	if installerOptions.InstallNamespace == "" {
		installerOptions.InstallNamespace = cliconstants.DefaultIstioOperatorNamespace
	}
}
