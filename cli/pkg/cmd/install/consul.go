package install

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/constants"
)

func installConsul(opts *options.Options) {

	installClient, err := getInstallClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	installSpec := generateInstallSpecFromOpts(opts)
	_, err = installClient.Write(installSpec, clients.WriteOpts{})
	if err != nil {
		fmt.Println(err)
		return
	}
	installationSummaryMessage(opts)
	return
}

func generateInstallSpecFromOpts(opts *options.Options) *v1.Install {
	installSpec := &v1.Install{
		Metadata: core.Metadata{
			Name:      getNewInstallName(opts),
			Namespace: constants.SuperglooNamespace,
		},
		Consul: &v1.ConsulInstall{
			Path:      constants.ConsulInstallPath,
			Namespace: opts.Install.Namespace,
		},
	}
	if opts.Install.Mtls {
		installSpec.Encryption = &v1.Encryption{
			TlsEnabled: opts.Install.Mtls,
			Secret:     &opts.Install.SecretRef,
		}
	}
	return installSpec
}
