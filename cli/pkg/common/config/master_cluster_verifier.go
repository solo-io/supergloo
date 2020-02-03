package common_config

import (
	"os"

	"github.com/rotisserie/eris"
	"k8s.io/client-go/rest"
)

var (
	FailedToParseContext = func(err error) error {
		return eris.Wrap(err, "Could not parse target kube context information")
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
//go:generate mockgen -destination ../../mocks/mock_master_cluster_verifier.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common/config MasterKubeConfigVerifier
type MasterKubeConfigVerifier interface {
	Verify(masterKubeConfigPath string, masterContext string) (err error)
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

func (m *masterKubeConfigVerifier) Verify(masterKubeConfigPath string, masterContext string) (err error) {
	if masterKubeConfigPath == "" {
		return MustProvideMasterConfigPath
	}

	fileExists, err := m.fileExistenceChecker(masterKubeConfigPath)
	if err != nil {
		return FailedToCheckFileExistence(err, masterKubeConfigPath)
	}
	if !fileExists {
		return FileDoesNotExist(masterKubeConfigPath)
	}

	_, err = m.kubeLoader.GetRawConfigForContext(masterKubeConfigPath, masterContext)
	if err != nil {
		return FailedToParseContext(err)
	}
	isSMH, err := m.verifyIsMaster()
	if err != nil {
		return CouldNotVerifyMaster(err, masterKubeConfigPath)
	}
	if !isSMH {
		return NoSmhInstallationFound(masterKubeConfigPath)
	}
	return nil
}

func (m *masterKubeConfigVerifier) verifyIsMaster() (bool, error) {
	//TODO: Implement this check for real. Look for mesh discovery or something ¯\_(ツ)_/¯
	return true, nil
}
