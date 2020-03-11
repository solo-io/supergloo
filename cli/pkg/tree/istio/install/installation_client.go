package install

import (
	"fmt"
	"io"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/common/kube"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator"
)

var (
	ConflictingControlPlaneSettings   = eris.New("Cannot both use a pre-configured Istio profile and provide an IstioControlPlane Custom Resource")
	FailedToParseControlPlaneSettings = func(err error) error {
		return eris.Wrap(err, "Failed to parse the provided IstioControlPlane resource")
	}
	FailedToParseControlPlaneWithProfile = func(err error, profile string) error {
		return eris.Wrapf(err, "Failed to parse the pre-configured IstioControlPlane with profile %s", profile)
	}
	FailedToWriteControlPlane = func(err error) error {
		return eris.Wrap(err, "Failed to write the provided IstioControlPlane resource")
	}
	TooManyControlPlaneResources = func(numResources int) error {
		return eris.Errorf("Expected the IstioControlPlane manifest to have only a single resource, found %d", numResources)
	}
	UnknownControlPlaneKind = func(kind string) error {
		return eris.Errorf("Expected the manifest to contain an IstioControlPlane, but found %s", kind)
	}

	istioControlPlaneKind = "IstioControlPlane"
)

type IstioInstaller interface {
	Install() error
}

func NewIstioInstaller(
	unstructuredKubeClient kube.UnstructuredKubeClient,
	istioInstallOptions *options.IstioInstall,
	clusterName string,
	out io.Writer,
	manifestBuilder operator.InstallerManifestBuilder,
	operatorManager operator.OperatorManager,
	fileReader common.FileReader,
) IstioInstaller {

	return &istioInstaller{
		unstructuredKubeClient: unstructuredKubeClient,
		manifestBuilder:        manifestBuilder,
		istioInstallOptions:    istioInstallOptions,
		out:                    out,
		clusterName:            clusterName,
		operatorManager:        operatorManager,
		fileReader:             fileReader,
	}
}

type istioInstaller struct {
	unstructuredKubeClient kube.UnstructuredKubeClient
	manifestBuilder        operator.InstallerManifestBuilder
	istioInstallOptions    *options.IstioInstall
	out                    io.Writer
	clusterName            string
	operatorManager        operator.OperatorManager
	fileReader             common.FileReader
}

func (i *istioInstaller) Install() error {
	namespace := i.istioInstallOptions.InstallationConfig.InstallNamespace

	istioControlPlane, err := i.loadIstioControlPlane()
	if err != nil {
		return err
	}

	if i.istioInstallOptions.DryRun {
		manifest, err := i.manifestBuilder.Build(&i.istioInstallOptions.InstallationConfig)
		if err != nil {
			return err
		}

		fmt.Fprintln(i.out, manifest)

		if istioControlPlane != "" {
			fmt.Fprintf(i.out, "\n---\n%s\n", istioControlPlane)
		}
		return nil
	}

	err = i.installOperator(namespace)
	if err != nil {
		return err
	}

	err = i.writeControlPlaneResource(namespace, istioControlPlane)
	if err != nil {
		return err
	}

	return nil
}

func (i *istioInstaller) installOperator(namespace string) error {
	installNeeded, err := i.operatorManager.ValidateOperatorNamespace(i.clusterName)
	if err != nil {
		return eris.Wrapf(err, "Istio operator namespace validation failed for cluster '%s' in namespace '%s'", i.clusterName, namespace)
	}

	// install the operator if it didn't exist already
	if installNeeded {
		fmt.Fprintf(i.out, "Installing the Istio operator to cluster '%s' in namespace '%s'\n", i.clusterName, namespace)

		err := i.operatorManager.Install()
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(i.out, "The Istio operator is already installed to cluster '%s' in namespace '%s' and is suitable for use. Continuing with the Istio installation.\n", i.clusterName, namespace)
	}

	return nil
}

func (i *istioInstaller) loadIstioControlPlane() (string, error) {
	userPath := i.istioInstallOptions.IstioControlPlaneManifestPath
	profile := i.istioInstallOptions.Profile

	if userPath != "" && profile != "" {
		return "", ConflictingControlPlaneSettings
	}

	if userPath != "" {
		userSpecifiedControlPlane, err := i.loadControlPlaneFromUserFlagConfig()
		if err != nil {
			return "", err
		}
		return userSpecifiedControlPlane, nil
	} else if profile != "" {
		preConfiguredProfile, err := i.manifestBuilder.GetControlPlaneSpecWithProfile(profile, i.istioInstallOptions.InstallationConfig.InstallNamespace)
		if err != nil {
			return "", FailedToParseControlPlaneWithProfile(err, profile)
		}
		return preConfiguredProfile, nil
	}

	return "", nil
}

// returns "", nil if the user did not provide an IstioControlPlane
func (i *istioInstaller) loadControlPlaneFromUserFlagConfig() (string, error) {
	path := i.istioInstallOptions.IstioControlPlaneManifestPath

	fileExists, err := i.fileReader.Exists(path)
	if err != nil {
		return "", eris.Wrapf(err, "Unexpected error while reading IstioControlPlane spec")
	} else if !fileExists {
		return "", eris.Errorf("Path to IstioControlPlane spec does not exist: %s", i.istioInstallOptions.IstioControlPlaneManifestPath)
	}

	contents, err := i.fileReader.Read(path)
	if err != nil {
		return "", err
	}

	stringContents := string(contents)
	return stringContents, nil
}

// the userProvidedControlPlane may be nil if they didn't provide one
func (i *istioInstaller) writeControlPlaneResource(namespace, istioControlPlaneToWrite string) error {
	if istioControlPlaneToWrite == "" {
		fmt.Fprintf(i.out,
			"\nThe Istio operator has been installed to cluster '%s' in namespace '%s'. No IstioControlPlane custom resource was provided to meshctl, so Istio is currently not fully installed yet. Write a IstioControlPlane CR to cluster '%s' to complete your installation\n",
			i.clusterName,
			namespace,
			i.clusterName,
		)
		return nil
	}

	// write the control plane
	resources, err := i.unstructuredKubeClient.BuildResources(namespace, istioControlPlaneToWrite)
	if err != nil {
		return FailedToParseControlPlaneSettings(err)
	}

	if len(resources) != 1 {
		return TooManyControlPlaneResources(len(resources))
	}

	istioControlPlane := resources[0]
	resourceKind := istioControlPlane.Object.GetObjectKind().GroupVersionKind().Kind
	if resourceKind != istioControlPlaneKind {
		return UnknownControlPlaneKind(resourceKind)
	}

	_, err = i.unstructuredKubeClient.Create(namespace, resources)
	if err != nil {
		return FailedToWriteControlPlane(err)
	}

	fmt.Fprintf(i.out, "\nThe IstioControlPlane has been written to cluster '%s' in namespace '%s'. The Istio operator should process it momentarily and install Istio.\n", i.clusterName, namespace)
	return nil
}
