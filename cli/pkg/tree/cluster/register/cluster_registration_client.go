package register

import (
	"fmt"

	"io"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	FailedLoadingTargetConfig = func(err error) error {
		return eris.Wrap(err, "Failed to load the kube config for the target cluster")
	}
	FailedToCreateAuthToken = func(err error, saRef *core.ResourceRef, targetKubeConfig string) error {
		return eris.Wrapf(err, "Failed to create an auth token for service account %s.%s in cluster pointed to by kube config %s. This operation is not atomic, so the service account may have been created and left in the cluster while a later step failed.", saRef.Namespace, saRef.Name, targetKubeConfig)
	}
	FailedToConvertToSecret = func(err error) error {
		return eris.Wrap(err, "Could not convert kube config for new service account into secret")
	}
	FailedToWriteSecret = func(err error) error {
		return eris.Wrap(err, "Could not write secret to master cluster")
	}
)

// write a new kube config secret to the master cluster containing creds for talking to the target cluster as the given service account
type ClusterRegistrationClient interface {
	RegisterCluster(cmd *cobra.Command, args []string) error
}

// most of these references will be invalid when this constructor runs- we just care about
// being able to get to these values at game time (when the command actually runs and flags
// have been parsed)
func NewClusterRegistrationClient(out io.Writer, flagConfig *cluster.FlagConfig, clientsFactory common.ClientsFactory) ClusterRegistrationClient {
	return &clusterRegistrationClient{
		out:            out,
		flagConfig:     flagConfig,
		clientsFactory: clientsFactory,
	}
}

type clusterRegistrationClient struct {
	out            io.Writer
	flagConfig     *cluster.FlagConfig
	clientsFactory common.ClientsFactory
}

func (c *clusterRegistrationClient) RegisterCluster(cmd *cobra.Command, args []string) error {
	// game time- build the clients and use the values
	clients, err := c.clientsFactory(c.flagConfig.GlobalFlagConfig.MasterKubeConfig, c.flagConfig.GlobalFlagConfig.MasterWriteNamespace)
	if err != nil {
		return err
	}

	// first we need the credentials to talk to the target cluster
	targetAuthConfig, err := clients.KubeLoader.GetRestConfig(c.flagConfig.TargetKubeConfigPath)
	if err != nil {
		return FailedLoadingTargetConfig(err)
	}

	// the new cluster name doubles as the name for the service account we will auth as
	serviceAccountRef := &core.ResourceRef{
		Name:      c.flagConfig.TargetClusterName,
		Namespace: c.flagConfig.TargetWriteNamespace,
	}
	configForServiceAccount, err := clients.ClusterAuthorization.CreateAuthConfigForCluster(targetAuthConfig, serviceAccountRef)
	if err != nil {
		return FailedToCreateAuthToken(err, serviceAccountRef, c.flagConfig.TargetKubeConfigPath)
	}

	// now we need the cluster/context information from that config
	ctx, err := clients.KubeLoader.ParseContext(c.flagConfig.TargetKubeConfigPath)
	if err != nil {
		return common.FailedToParseContext(err)
	}

	targetCurrentContext := ctx.CurrentContext
	targetContext := ctx.Contexts[targetCurrentContext]
	targetCluster := ctx.Clusters[targetContext.Cluster]

	fmt.Fprintf(c.out, "Successfully wrote service account to target cluster...\n")

	secret, err := kubeconfig.KubeConfigToSecret(
		c.flagConfig.TargetClusterName,
		c.flagConfig.GlobalFlagConfig.MasterWriteNamespace,
		&kubeconfig.KubeConfig{
			Config: api.Config{
				Kind:        "Secret",
				APIVersion:  "v1",
				Preferences: api.Preferences{},
				Clusters: map[string]*api.Cluster{
					c.flagConfig.TargetClusterName: targetCluster,
				},
				AuthInfos: map[string]*api.AuthInfo{
					c.flagConfig.TargetClusterName: {
						Token: configForServiceAccount.BearerToken,
					},
				},
				Contexts: map[string]*api.Context{
					c.flagConfig.TargetClusterName: {
						LocationOfOrigin: targetContext.LocationOfOrigin,
						Cluster:          c.flagConfig.TargetClusterName,
						AuthInfo:         c.flagConfig.TargetClusterName,
						Namespace:        targetContext.Namespace,
						Extensions:       targetContext.Extensions,
					},
				},
				CurrentContext: c.flagConfig.TargetClusterName,
			},
			Cluster: c.flagConfig.TargetClusterName,
		})

	if err != nil {
		return FailedToConvertToSecret(err)
	}

	err = clients.SecretWriter.Write(secret)
	if err != nil {
		return FailedToWriteSecret(err)
	}

	fmt.Fprintf(c.out, "Successfully wrote kube config secret to master cluster...\n")
	fmt.Fprintf(c.out, "\nCluster %s is now registered in your Service Mesh Hub installation\n", c.flagConfig.TargetClusterName)

	return nil
}
