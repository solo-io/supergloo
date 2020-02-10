package install

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/spf13/cobra"
)

type IstioInstallationCmd *cobra.Command

var IstioInstallationProviderSet = wire.NewSet(
	BuildIstioInstallationCmd,
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

	flags := cmd.PersistentFlags()

	operatorNsFlag := "operator-namespace"
	profilesUsage := fmt.Sprintf(
		"Install Istio in one of its pre-configured profiles; supported profiles: [%s] (https://preliminary.istio.io/docs/setup/additional-setup/config-profiles/)",
		strings.Join(operator.ValidProfiles.List(), ", "),
	)

	flags.StringVar(&opts.Istio.Install.InstallationConfig.IstioOperatorVersion, "operator-version", operator.DefaultIstioOperatorVersion, "Version of the Istio operator to use (https://hub.docker.com/r/istio/operator/tags)")
	flags.StringVar(&opts.Istio.Install.InstallationConfig.InstallNamespace, operatorNsFlag, operator.DefaultIstioOperatorNamespace, "Namespace in which to install the Istio operator")
	flags.BoolVar(&opts.Istio.Install.InstallationConfig.CreateIstioControlPlaneCRD, "create-operator-crd", true, "Register the IstioControlPlane CRD in the target cluster")
	flags.BoolVar(&opts.Istio.Install.InstallationConfig.CreateNamespace, "create-operator-namespace", true, "Create the namespace specified by --"+operatorNsFlag)
	flags.BoolVar(&opts.Istio.Install.DryRun, "dry-run", false, "Dump the manifest that would be used to install the operator to stdout rather than apply it")
	flags.StringVar(&opts.Istio.Install.IstioControlPlaneManifestPath, "control-plane-spec", "", "Optional path to a YAML file containing an IstioControlPlane resource")
	flags.StringVar(&opts.Istio.Install.Profile, "profile", "", profilesUsage)

	options.AddRootFlags(flags, opts)

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

	restClientGetter := kubeLoader.RESTClientGetter(kubeConfigPath)
	unstructuredKubeClient := clients.UnstructuredKubeClientFactory(restClientGetter)

	rawCfg, err := kubeLoader.GetRawConfigForContext(kubeConfigPath, kubeContext)
	if err != nil {
		return nil, err
	}

	contextName := kubeContext
	if contextName == "" {
		contextName = rawCfg.CurrentContext
	}
	clusterName := rawCfg.Contexts[contextName].Cluster

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
