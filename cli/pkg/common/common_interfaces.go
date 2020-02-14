package common

import (
	"io/ioutil"
	"os"

	"github.com/solo-io/go-utils/installutils/helminstall"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/common/kube"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	k8sapiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination ../mocks/mock_common_interfaces.go -package cli_mocks -source ./common_interfaces.go

// a grab bag of various clients that command implementations may use
type KubeClients struct {
	ClusterAuthorization auth.ClusterAuthorization
	SecretWriter         SecretWriter
	HelmInstaller        helminstall.Installer
	KubeClusterClient    discovery_core.KubernetesClusterClient // client for KubernetesCluster custom resources
}

type KubeClientsFactory func(masterConfig *rest.Config, writeNamespace string) (*KubeClients, error)

type Clients struct {
	ServerVersionClient           server.ServerVersionClient
	MasterClusterVerifier         common_config.MasterKubeConfigVerifier
	ReleaseAssetHelper            upgrade_assets.AssetHelper
	UnstructuredKubeClientFactory kube.UnstructuredKubeClientFactory
	DeploymentClient              server.DeploymentClient

	IstioClients IstioClients
}

func IstioClientsProvider(manifestBuilder operator.InstallerManifestBuilder, operatorManagerFactory operator.OperatorManagerFactory) IstioClients {
	return IstioClients{
		OperatorManifestBuilder: manifestBuilder,
		OperatorManagerFactory:  operatorManagerFactory,
	}
}

type IstioClients struct {
	OperatorManifestBuilder operator.InstallerManifestBuilder
	OperatorManagerFactory  operator.OperatorManagerFactory
}

type ClientsFactory func(opts *options.Options) (*Clients, error)

// write a k8s secret
// used to pare down from the complexity of all the methods on the k8s client-go SecretInterface
type SecretWriter interface {
	// create, or update if already exists
	Apply(secret *k8sapiv1.Secret) error
}

func ClientsProvider(
	serverVersionClient server.ServerVersionClient,
	assetHelper upgrade_assets.AssetHelper,
	verifier common_config.MasterKubeConfigVerifier,
	unstructuredKubeClientFactory kube.UnstructuredKubeClientFactory,
	deploymentClient server.DeploymentClient,
	istioClients IstioClients,
) *Clients {
	return &Clients{
		ServerVersionClient:           serverVersionClient,
		MasterClusterVerifier:         verifier,
		UnstructuredKubeClientFactory: unstructuredKubeClientFactory,
		DeploymentClient:              deploymentClient,
		ReleaseAssetHelper:            assetHelper,
		IstioClients:                  istioClients,
	}
}

// facilitates wire codegen
func KubeClientsProvider(
	authorization auth.ClusterAuthorization,
	writer SecretWriter,
	helmInstaller helminstall.Installer,
	kubeClusterClient discovery_core.KubernetesClusterClient,
) *KubeClients {
	return &KubeClients{
		ClusterAuthorization: authorization,
		SecretWriter:         writer,
		HelmInstaller:        helmInstaller,
		KubeClusterClient:    kubeClusterClient,
	}
}

type FileReader interface {
	Read(filePath string) ([]byte, error)
	Exists(filePath string) (exists bool, err error)
}

func NewDefaultFileReader() FileReader {
	return &fileReader{}
}

type fileReader struct{}

func (f *fileReader) Read(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func (f *fileReader) Exists(filePath string) (exists bool, err error) {
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
