package operator

import (
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	k8s_apps "k8s.io/api/apps/v1"
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
	FailedToValidateExistingOperator = func(err error, namespace string) error {
		return eris.Wrapf(err, "Failed to validate existing Mesh operator deployment in namespace %s", namespace)
	}
	FailedToDetermineOperatorVersion = func(current string, minimumSupportedVersion versionutils.Version) error {
		return eris.Errorf("Failed to determine whether the current operator running at version %s is the minimum version %s", current, minimumSupportedVersion.String())
	}
	IncompatibleOperatorVersion = func(current string, minimumSupportedVersion string) error {
		return eris.Errorf("Found istio operator running at version %s, while %s is the minimum supported version", current, minimumSupportedVersion)
	}
	FailedToCheckIfOperatorExists = func(err error, namespace string) error {
		return eris.Wrapf(err, "Failed to check whether the Mesh operator is already installed in namespace %s", namespace)
	}

	istioVersionToMinimumOperatorVersion = map[IstioVersion]versionutils.Version{
		IstioVersion1_5: {
			Major: 1,
			Minor: 5,
			Patch: 1,
		},
		IstioVersion1_6: {
			Major: 1,
			Minor: 6,
			Patch: 0,
		},
	}
)

type OperatorManagerFactory func(
	operatorDao OperatorDao,
	manifestBuilder InstallerManifestBuilder,
	imageNameParser docker.ImageNameParser,
) OperatorManager

func NewOperatorManagerFactory() OperatorManagerFactory {
	return NewManager
}

func NewManager(
	operatorDao OperatorDao,
	manifestBuilder InstallerManifestBuilder,
	imageNameParser docker.ImageNameParser,
) OperatorManager {
	return &manager{
		operatorDao:     operatorDao,
		manifestBuilder: manifestBuilder,
		imageNameParser: imageNameParser,
	}
}

type manager struct {
	manifestBuilder InstallerManifestBuilder
	operatorDao     OperatorDao
	imageNameParser docker.ImageNameParser
}

func (m *manager) InstallOperatorApplication(installationOptions *InstallationOptions) error {
	namespace := m.getNamespaceFromOptions(installationOptions)
	manifest, err := m.InstallDryRun(installationOptions)
	if err != nil {
		return err
	}

	installNeeded, err := m.validateOperatorNamespace(installationOptions.IstioVersion, DefaultIstioOperatorDeploymentName, DefaultIstioOperatorDeploymentName)
	if err != nil {
		return err
	}

	if installNeeded {
		err = m.operatorDao.ApplyManifest(namespace, manifest)
		if err != nil {
			return FailedToInstallOperator(err)
		}
	}

	return m.writeOperatorConfig(installationOptions)
}

func (m *manager) InstallDryRun(installationOptions *InstallationOptions) (manifest string, err error) {
	namespace := m.getNamespaceFromOptions(installationOptions)

	applicationManifest, err := m.manifestBuilder.BuildOperatorDeploymentManifest(installationOptions.IstioVersion, namespace, installationOptions.CreateNamespace)
	if err != nil {
		return "", FailedToGenerateInstallManifest(err)
	}

	return applicationManifest, nil
}

func (m *manager) OperatorConfigDryRun(installationOptions *InstallationOptions) (manifest string, err error) {
	namespace := m.getNamespaceFromOptions(installationOptions)

	var operatorConfig string
	if installationOptions.InstallationProfile != "" {
		operatorConfig, err = m.manifestBuilder.BuildOperatorConfigurationWithProfile(installationOptions.InstallationProfile, namespace)
		if err != nil {
			return "", err
		}
	} else if installationOptions.CustomIstioOperatorSpec != "" {
		operatorConfig = installationOptions.CustomIstioOperatorSpec
	} else {
		operatorConfig, err = m.manifestBuilder.BuildOperatorConfigurationWithProfile(DefaultIstioProfile, namespace)
		if err != nil {
			return "", err
		}
	}

	return operatorConfig, nil
}

func (m *manager) writeOperatorConfig(installationOptions *InstallationOptions) error {
	configManifest, err := m.OperatorConfigDryRun(installationOptions)
	if err != nil {
		return err
	}

	err = m.operatorDao.ApplyManifest(m.getNamespaceFromOptions(installationOptions), configManifest)
	if err != nil {
		return FailedToInstallOperator(err)
	}

	return nil
}

// ensure that the given namespace is appropriate for installing an Mesh operator
// will fail with an error if the operator is already present at a different version than we're specifying
// will return (true, nil) if no operator deployment is present yet
// (false, nil) indicates the operator is present already at an appropriate version, so no need to call .Install()
//
// the `clusterName` arg is only used for error reporting
func (m *manager) validateOperatorNamespace(istioVersion IstioVersion, operatorName, operatorNamespace string) (installNeeded bool, err error) {
	operatorDeployment, err := m.operatorDao.FindOperatorDeployment(operatorName, operatorNamespace)
	if err != nil {
		return false, FailedToCheckIfOperatorExists(err, operatorNamespace)
	} else if operatorDeployment == nil {
		return true, nil
	}

	err = m.validateExistingOperatorDeployment(istioVersion, operatorDeployment)
	if err != nil {
		return false, FailedToValidateExistingOperator(err, operatorNamespace)
	}

	return false, nil
}

func (*manager) getNamespaceFromOptions(options *InstallationOptions) string {
	namespace := options.InstallNamespace
	if namespace == "" {
		namespace = DefaultIstioOperatorNamespace
	}

	return namespace
}

func (m *manager) validateExistingOperatorDeployment(istioVersion IstioVersion, deployment *k8s_apps.Deployment) error {
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

	minimumSupportedVersion, ok := istioVersionToMinimumOperatorVersion[istioVersion]
	if !ok {
		return eris.Errorf("Istio version %s does not have a minimum supported version; this is unexpected", string(istioVersion))
	}
	greater, ok := version.IsGreaterThan(minimumSupportedVersion)
	if !ok {
		return FailedToDetermineOperatorVersion(version.String(), minimumSupportedVersion)
	}

	if !(greater || version.Equals(&minimumSupportedVersion)) {
		return IncompatibleOperatorVersion(version.String(), minimumSupportedVersion.String())
	}

	return nil
}
