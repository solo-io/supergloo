package register

import (
	"fmt"
	"io"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	cluster_common "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/common"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	kubev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
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
	FailedToWriteKubeCluster = func(err error) error {
		return eris.Wrap(err, "Could not write KubernetesCluster resource to master cluster")
	}
)

// write a new kube config secret to the master cluster containing creds for talking to the target cluster as the given service account

func RegisterCluster(
	clientsFactory common.ClientsFactory,
	kubeClientsFactory common.KubeClientsFactory,
	opts *options.Options,
	out io.Writer,
	kubeLoader common_config.KubeLoader,
) error {

	if err := cluster_common.VerifyRemoteContextFlags(opts); err != nil {
		return err
	}
	if err := cluster_common.VerifyMasterCluster(clientsFactory, opts); err != nil {
		return err
	}

	// first we need the credentials to talk to the target cluster
	cfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
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

	configForServiceAccount, err := generateServiceAccountConfig(out, kubeClients, kubeLoader, remoteKubeConfig, registerOpts)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Successfully wrote service account to target cluster...\n")

	secret, err := writeKubeConfigToMaster(
		opts.Root.WriteNamespace,
		registerOpts,
		remoteKubeConfig,
		configForServiceAccount,
		kubeClients,
		kubeLoader,
	)
	if err != nil {
		return err
	}

	err = writeKubeCluster(kubeClients, opts.Root.WriteNamespace, registerOpts, secret)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Successfully wrote kube config secret to master cluster...\n")
	fmt.Fprintf(out, "\nCluster %s is now registered in your Service Mesh Hub installation\n", registerOpts.RemoteClusterName)

	return nil
}

func generateServiceAccountConfig(
	out io.Writer,
	kubeClients *common.KubeClients,
	kubeLoader common_config.KubeLoader,
	remoteKubeConfig string,
	registerOpts options.Register,
) (*rest.Config, error) {

	// first we need the credentials to talk to the target cluster
	targetAuthConfig, err := kubeLoader.GetRestConfigForContext(remoteKubeConfig, registerOpts.RemoteContext)
	if err != nil {
		return nil, FailedLoadingRemoteConfig(err)
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
		return nil, err
	}

	return configForServiceAccount, nil
}

func writeKubeConfigToMaster(
	writeNamespace string,
	registerOpts options.Register,
	remoteKubeConfig string,
	serviceAccountConfig *rest.Config,
	kubeClients *common.KubeClients,
	kubeLoader common_config.KubeLoader,
) (*kubev1.Secret, error) {

	// now we need the cluster/context information from that config
	kubeCtx, err := kubeLoader.GetRawConfigForContext(remoteKubeConfig, registerOpts.RemoteContext)
	if err != nil {
		return nil, common_config.FailedToParseContext(err)
	}

	remoteContext := kubeCtx.CurrentContext
	if registerOpts.RemoteContext != "" {
		remoteContext = registerOpts.RemoteContext
	}

	targetContext := kubeCtx.Contexts[remoteContext]
	targetCluster := kubeCtx.Clusters[targetContext.Cluster]

	secret, err := kubeconfig.KubeConfigToSecret(
		registerOpts.RemoteClusterName,
		writeNamespace,
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
						Token: serviceAccountConfig.BearerToken,
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
		return nil, FailedToConvertToSecret(err)
	}

	err = kubeClients.SecretWriter.Apply(secret)
	if err != nil {
		return nil, FailedToWriteSecret(err)
	}

	return secret, nil
}

// write the KubernetesCluster resource to the master cluster
func writeKubeCluster(
	kubeClients *common.KubeClients,
	writeNamespace string,
	registerOpts options.Register,
	secret *kubev1.Secret,
) error {

	err := kubeClients.KubeClusterClient.Create(&v1alpha1.KubernetesCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      registerOpts.RemoteClusterName,
			Namespace: writeNamespace,
		},
		Spec: types.KubernetesClusterSpec{
			SecretRef: &types.ResourceRef{
				Name:      secret.GetName(),
				Namespace: secret.GetNamespace(),
			},
		},
	})
	if err != nil {
		return FailedToWriteKubeCluster(err)
	}
	return nil
}
