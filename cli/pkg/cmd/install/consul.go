package install

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/constants"
)

func installConsul(opts *options.Options) {

	cfg, err := kubeutils.GetConfig("", "")
	cache := kube.NewKubeCache()
	if err != nil {
		fmt.Println(err)
		return
	}
	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	_, err = installClient.Write(&v1.Install{
		Metadata: core.Metadata{
			Name:      getNewInstallName(opts),
			Namespace: constants.SuperGlooNamespace,
		},
		Consul: &v1.ConsulInstall{
			Path:      constants.ConsulInstallPath,
			Namespace: opts.Install.Namespace,
		}}, clients.WriteOpts{})
	if err != nil {
		fmt.Println(err)
		return
	}
	installationSummaryMessage(opts)
	return
}
