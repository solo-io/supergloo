package istio

import (
	"context"
	"fmt"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/filesystem/files"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
	"github.com/spf13/cobra"
)

func BuildIstioInstallationRunFunc(
	ctx context.Context,
	out io.Writer,
	istioVersion operator.IstioVersion,
	kubeLoader kubeconfig.KubeLoader,
	opts *options.Options,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	imageNameParser docker.ImageNameParser,
	fileReader files.FileReader,
) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
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

		istioInstallOpts := &operator.InstallationOptions{
			IstioVersion:            istioVersion,
			InstallNamespace:        opts.Mesh.Install.InstallationConfig.InstallNamespace,
			CreateNamespace:         opts.Mesh.Install.InstallationConfig.CreateNamespace,
			InstallationProfile:     opts.Mesh.Install.Profile,
			CustomIstioOperatorSpec: customSpec,
		}
		if opts.Mesh.Install.DryRun {
			manifest, err := operatorManager.OperatorConfigDryRun(istioInstallOpts)
			if err != nil {
				return err
			}

			fmt.Fprintln(out, manifest)
			return nil
		} else {
			return operatorManager.InstallOperatorApplication(istioInstallOpts)
		}
	}
}
