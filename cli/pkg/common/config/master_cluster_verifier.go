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
		return eris.Errorf("Could not find a Service Mesh Hub installation in the cluster pointed "+
			"to by the kube config %s", path)
	}
	CouldNotVerifyMaster = func(err error, path string) error {
		return eris.Wrapf(err, "Could not verify whether a Service Mesh Hub installation is present "+
			"in the cluster pointed to by the kube config %s", path)
	}
	FileDoesNotExist = func(err error, path string) error {
		return eris.Wrapf(err, "Kube config at %s does not exist", path)
	}
)

// Verify that the cluster pointed to by the given kube config is actually a Service Mesh Hub installation
//go:generate mockgen -destination ../../mocks/mock_master_cluster_verifier.go -package cli_mocks github.com/solo-io/mesh-projects/cli/pkg/common/config MasterKubeConfigVerifier
type MasterKubeConfigVerifier interface {
	Verify(masterKubeConfigPath string, masterContext string) (err error)
}

// Accepts the validated master kube REST config
type OnMasterVerificationSuccess func(masterKubeConfig *rest.Config)

func NewMasterKubeConfigVerifier(kubeLoader KubeLoader) MasterKubeConfigVerifier {
	return &masterKubeConfigVerifier{
		kubeLoader: kubeLoader,
	}
}

type masterKubeConfigVerifier struct {
	kubeLoader KubeLoader
}

func (m *masterKubeConfigVerifier) Verify(masterKubeConfigPath string, masterContext string) (err error) {
	_, err = m.kubeLoader.GetRawConfigForContext(masterKubeConfigPath, masterContext)
	if err != nil {
		if os.IsNotExist(err) {
			return FileDoesNotExist(err, masterKubeConfigPath)
		}
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
