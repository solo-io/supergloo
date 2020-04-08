package cluster_internal

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
)

var (
	NoRemoteConfigSpecifiedError = eris.New("flag(s) missing: must set either --remote-kubeconfig or " +
		"--remote-context")
)

func VerifyMasterCluster(factory common.ClientsFactory, opts *options.Options) error {
	clients, err := factory(opts)
	if err != nil {
		return err
	}
	err = clients.MasterClusterVerifier.Verify(opts.Root.KubeConfig, opts.Root.KubeContext)
	return err
}

func VerifyRemoteContextFlags(opts *options.Options) error {
	if opts.Cluster.Register.RemoteKubeConfig == "" &&
		opts.Cluster.Register.RemoteContext == "" {
		return NoRemoteConfigSpecifiedError
	}
	return nil
}
