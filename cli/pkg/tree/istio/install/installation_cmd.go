package install

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/spf13/cobra"
)

type IstioInstallationCmd *cobra.Command

var (
	IstioInstallationProviderSet = wire.NewSet(
		BuildIstioInstallationCmd,
	)
	ContextNotFound = func(contextName string) error {
		return eris.Errorf("Context '%s' not found in kubeconfig file", contextName)
	}
)

func BuildIstioInstallationCmd(
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	out io.Writer,
	kubeLoader common_config.KubeLoader,
	imageNameParser docker.ImageNameParser,
	fileReader common.FileReader,
) IstioInstallationCmd {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Istio on the indicated cluster using the Istio installation operator (https://preliminary.istio.io/docs/setup/install/standalone-operator/)",
		RunE: func(cmd *cobra.Command, args []string) error {
			clients, err := clientsFactory(opts)
			if err != nil {
				return err
			}

			istioInstaller, err := buildIstioInstaller(out, clients, opts, opts.Root.KubeConfig, opts.Root.KubeContext, kubeLoader, imageNameParser, fileReader)
			if err != nil {
				return err
			}

			return istioInstaller.Install()
		},
	}
	profilesUsage := fmt.Sprintf(
		"Install Istio in one of its pre-configured profiles; supported profiles: [%s] (https://preliminary.istio.io/docs/setup/additional-setup/config-profiles/)",
		strings.Join(operator.ValidProfiles.List(), ", "),
	)
	options.AddIstioInstallFlags(cmd, opts, profilesUsage)
	return cmd
}

func buildIstioInstaller(
	out io.Writer,
	clients *common.Clients,
	opts *options.Options,
	kubeConfigPath,
	kubeContext string,
	kubeLoader common_config.KubeLoader,
	imageNameParser docker.ImageNameParser,
	fileReader common.FileReader,
) (IstioInstaller, error) {

	restClientGetter := kubeLoader.RESTClientGetter(kubeConfigPath, kubeContext)
	unstructuredKubeClient := clients.UnstructuredKubeClientFactory(restClientGetter)

	rawCfg, err := kubeLoader.GetRawConfigForContext(kubeConfigPath, kubeContext)
	if err != nil {
		return nil, err
	}

	contextName := kubeContext
	if contextName == "" {
		contextName = rawCfg.CurrentContext
	}
	context, ok := rawCfg.Contexts[contextName]
	if !ok {
		return nil, ContextNotFound(contextName)
	}
	clusterName := context.Cluster

	operatorManager := clients.IstioClients.OperatorManagerFactory(
		unstructuredKubeClient,
		clients.IstioClients.OperatorManifestBuilder,
		clients.DeploymentClient,
		imageNameParser,
		&opts.Istio.Install.InstallationConfig,
	)

	return NewIstioInstaller(
		unstructuredKubeClient,
		&opts.Istio.Install,
		clusterName,
		out,
		clients.IstioClients.OperatorManifestBuilder,
		operatorManager,
		fileReader,
	), nil
}
