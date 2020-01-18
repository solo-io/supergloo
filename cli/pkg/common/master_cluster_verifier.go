package common

import (
	"os"

	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

var (
	FailedToParseContext = func(err error) error {
		return eris.Wrap(err, "Could not parse target kube context")
	}
	NoSmhInstallationFound = func(path string) error {
		return eris.Errorf("Could not find a Service Mesh Hub installation in the cluster pointed to by the kube config %s", path)
	}
	CouldNotVerifyMaster = func(err error, path string) error {
		return eris.Wrapf(err, "Could not verify whether a Service Mesh Hub installation is present in the cluster pointed to by the kube config %s", path)
	}
	MustProvideMasterConfigPath = eris.New("Must provide a path to the kube config for your master cluster by either providing the --master-cluster flag or setting the KUBECONFIG env var")
	FailedToCheckFileExistence  = func(err error, path string) error {
		return eris.Wrapf(err, "Failed to check whether the path %s exists", path)
	}
	FileDoesNotExist = func(path string) error {
		return eris.Errorf("Kube config at %s does not exist", path)
	}
)

// Verify that the cluster pointed to by the given kube config is actually a Service Mesh Hub installation
//go:generate mockgen -destination ../mocks/mock_master_cluster_verifier.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common MasterKubeConfigVerifier
type MasterKubeConfigVerifier interface {
	// Build and return a callback that cobra can run as a prerun step
	// If the cluster is determined to have a valid SMH installation, then `onSuccessCallback` is called with the parsed REST config for that cluster;
	// commonly used to return that rest config out of the verification step
	BuildVerificationCallback(masterKubeConfigPath *string, onSuccessCallback OnMasterVerificationSuccess) func(cmd *cobra.Command, args []string) (err error)
}

// Accepts the validated master kube REST config
type OnMasterVerificationSuccess func(masterKubeConfig *rest.Config)
type FileExistenceChecker func(path string) (bool, error)

func NewMasterKubeConfigVerifier(kubeLoader KubeLoader, fileExistenceChecker FileExistenceChecker) MasterKubeConfigVerifier {
	return &masterKubeConfigVerifier{
		kubeLoader:           kubeLoader,
		fileExistenceChecker: fileExistenceChecker,
	}
}

func DefaultFileExistenceCheckerProvider() FileExistenceChecker {
	return func(path string) (b bool, err error) {
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		return true, nil
	}
}

type masterKubeConfigVerifier struct {
	kubeLoader           KubeLoader
	fileExistenceChecker FileExistenceChecker
}

func (m *masterKubeConfigVerifier) BuildVerificationCallback(masterKubeConfigPath *string, onSuccessCallback OnMasterVerificationSuccess) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		if masterKubeConfigPath == nil || *masterKubeConfigPath == "" {
			return MustProvideMasterConfigPath
		}

		fileExists, err := m.fileExistenceChecker(*masterKubeConfigPath)
		if err != nil {
			return FailedToCheckFileExistence(err, *masterKubeConfigPath)
		}
		if !fileExists {
			return FileDoesNotExist(*masterKubeConfigPath)
		}

		_, err = m.kubeLoader.ParseContext(*masterKubeConfigPath)
		if err != nil {
			return FailedToParseContext(err)
		}
		isSMH, err := m.verifyIsMaster()
		if err != nil {
			return CouldNotVerifyMaster(err, *masterKubeConfigPath)
		}
		if !isSMH {
			return NoSmhInstallationFound(*masterKubeConfigPath)
		}

		masterKubeConfig, _ := m.kubeLoader.GetRestConfig(*masterKubeConfigPath)

		onSuccessCallback(masterKubeConfig)

		return nil
	}
}

func (m *masterKubeConfigVerifier) verifyIsMaster() (bool, error) {
	//TODO: Implement this check for real. Look for mesh discovery or something ¯\_(ツ)_/¯
	return true, nil
}
