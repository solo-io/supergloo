package register

import (
	"fmt"
	"io"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	cluster_common "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/common"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	FailedLoadingRemoteConfig = func(err error) error {
		return eris.Wrap(err, "Failed to load the kube config for the remote cluster")
	}
	FailedLoadingMasterConfig = func(err error) error {
		return eris.Wrap(err, "Failed to load the kube config for the master cluster")
	}
	FailedToCreateAuthToken = func(saRef *core.ResourceRef, remoteKubeConfig, remoteContext string) string {
		return fmt.Sprintf("Failed to create an auth token for service account %s.%s in cluster "+
			"pointed to by kube config %s with context %s. This operation is not atomic, so the service account may "+
			"have been created and left in the cluster while a later step failed. \n",
			saRef.Namespace, saRef.Name, remoteKubeConfig, remoteContext)
	}
	FailedToConvertToSecret = func(err error) error {
		return eris.Wrap(err, "Could not convert kube config for new service account into secret")
	}
	FailedToWriteSecret = func(err error) error {
		return eris.Wrap(err, "Could not write secret to master cluster")
	}
)

// write a new kube config secret to the master cluster containing creds for talking to the target cluster as the given service account

func RegisterCluster(
	clientsFactory common.ClientsFactory,
	kubeClientsFactory common.KubeClientsFactory,
	opts *options.Options,
	out io.Writer) error {

	if err := cluster_common.VerifyRootContextFlags(opts); err != nil {
		return err
	}
	if err := cluster_common.VerifyMasterCluster(clientsFactory, opts); err != nil {
		return err
	}

	clients, err := clientsFactory(opts)
	if err != nil {
		return err
	}
	// first we need the credentials to talk to the master cluster
	cfg, err := clients.KubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
	if err != nil {
		return FailedLoadingMasterConfig(err)
	}
	kubeClients, err := kubeClientsFactory(cfg, opts.Root.WriteNamespace)
	if err != nil {
		return err
	}

	registerOpts := opts.Cluster.Register

	remoteKubeConfig := opts.Root.KubeConfig
	if registerOpts.RemoteKubeConfig != "" {
		remoteKubeConfig = registerOpts.RemoteKubeConfig
	}

	// first we need the credentials to talk to the target cluster
	targetAuthConfig, err := clients.KubeLoader.GetRestConfigForContext(remoteKubeConfig, registerOpts.RemoteContext)
	if err != nil {
		return FailedLoadingRemoteConfig(err)
	}

	// the new cluster name doubles as the name for the service account we will auth as
	serviceAccountRef := &core.ResourceRef{
		Name:      registerOpts.RemoteClusterName,
		Namespace: registerOpts.RemoteWriteNamespace,
	}
	configForServiceAccount, err := kubeClients.ClusterAuthorization.
		CreateAuthConfigForCluster(targetAuthConfig, serviceAccountRef)
	if err != nil {
		fmt.Fprintf(out, FailedToCreateAuthToken(
			serviceAccountRef,
			registerOpts.RemoteKubeConfig,
			registerOpts.RemoteContext,
		))
		return err
	}

	// now we need the cluster/context information from that config
	ctx, err := clients.KubeLoader.GetRawConfigForContext(remoteKubeConfig, registerOpts.RemoteContext)
	if err != nil {
		return common_config.FailedToParseContext(err)
	}

	remoteContext := ctx.CurrentContext
	if registerOpts.RemoteContext != "" {
		remoteContext = registerOpts.RemoteContext
	}

	targetContext := ctx.Contexts[remoteContext]
	targetCluster := ctx.Clusters[targetContext.Cluster]

	fmt.Fprintf(out, "Successfully wrote service account to target cluster...\n")

	secret, err := kubeconfig.KubeConfigToSecret(
		registerOpts.RemoteClusterName,
		opts.Root.WriteNamespace,
		&kubeconfig.KubeConfig{
			Config: api.Config{
				Kind:        "Secret",
				APIVersion:  "v1",
				Preferences: api.Preferences{},
				Clusters: map[string]*api.Cluster{
					registerOpts.RemoteClusterName: targetCluster,
				},
				AuthInfos: map[string]*api.AuthInfo{
					registerOpts.RemoteClusterName: {
						Token: configForServiceAccount.BearerToken,
					},
				},
				Contexts: map[string]*api.Context{
					registerOpts.RemoteClusterName: {
						LocationOfOrigin: targetContext.LocationOfOrigin,
						Cluster:          registerOpts.RemoteClusterName,
						AuthInfo:         registerOpts.RemoteClusterName,
						Namespace:        targetContext.Namespace,
						Extensions:       targetContext.Extensions,
					},
				},
				CurrentContext: registerOpts.RemoteClusterName,
			},
			Cluster: registerOpts.RemoteClusterName,
		})

	if err != nil {
		return FailedToConvertToSecret(err)
	}

	err = kubeClients.SecretWriter.Apply(secret)
	if err != nil {
		return FailedToWriteSecret(err)
	}

	fmt.Fprintf(out, "Successfully wrote kube config secret to master cluster...\n")
	fmt.Fprintf(out, "\nCluster %s is now registered in your Service Mesh Hub installation\n",
		registerOpts.RemoteClusterName)

	return nil
}
