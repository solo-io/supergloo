package istio1_6

import (
	"context"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/filesystem/files"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
	"github.com/spf13/cobra"
)

type Istio1_6Cmd *cobra.Command

func NewIstio1_6InstallCmd(
	ctx context.Context,
	kubeLoader kubeconfig.KubeLoader,
	opts *options.Options,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	imageNameParser docker.ImageNameParser,
	fileReader files.FileReader,
) Istio1_6Cmd {
	cmd := &cobra.Command{
		Use:   cliconstants.Istio1_6Command.Use,
		Short: cliconstants.Istio1_6Command.Short,
		RunE: func(_ *cobra.Command, _ []string) error {
			clients, err := clientsFactory(opts)
			if err != nil {
				return err
			}
			restClientGetter := kubeLoader.RESTClientGetter(opts.Root.KubeConfig, opts.Root.KubeContext)
			cfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
			if err != nil {
				return err
			}

			kubeClients, err := kubeClientsFactory(cfg, opts.Root.WriteNamespace)
			if err != nil {
				return err
			}

			operatorManager := clients.IstioClients.OperatorManagerFactory(
				clients.IstioClients.OperatorDaoFactory(
					ctx,
					clients.UnstructuredKubeClientFactory(restClientGetter),
					kubeClients.DeploymentClient,
					imageNameParser,
				),
				clients.IstioClients.OperatorManifestBuilder,
				imageNameParser,
			)

			var customSpec string
			manifestPath := opts.Mesh.Install.ManifestPath
			if manifestPath != "" {
				if manifestPath == "-" {
					manifestPath = "/dev/stdin"
				}
				contents, err := fileReader.Read(manifestPath)
				if err != nil {
					return err
				}

				customSpec = string(contents)
			}

			return operatorManager.InstallOperatorApplication(&operator.InstallationOptions{
				IstioVersion:            operator.IstioVersion1_6,
				InstallNamespace:        opts.Mesh.Install.InstallationConfig.InstallNamespace,
				CreateNamespace:         opts.Mesh.Install.InstallationConfig.CreateNamespace,
				InstallationProfile:     opts.Mesh.Install.Profile,
				CustomIstioOperatorSpec: customSpec,
			})
		},
	}

	options.AddIstioInstallFlags(cmd, opts)

	return cmd
}
