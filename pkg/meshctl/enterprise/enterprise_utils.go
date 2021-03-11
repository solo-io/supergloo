package enterprise

import (
	"context"
	"strconv"

	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistrationOptions struct {
	registration.Options
	AgentChartPathOverride string
	AgentChartValuesPath   string

	RelayServerAddress  string
	RelayServerInsecure bool

	RootCASecretName      string
	RootCASecretNamespace string

	ClientCertSecretName      string
	ClientCertSecretNamespace string

	TokenSecretName      string
	TokenSecretNamespace string
	TokenSecretKey       string
}

func ensureCerts(ctx context.Context, opts RegistrationOptions) error {
	// if secure, and no user data was given, attempt to deduce the required parameters.
	const defaultRootCA = "relay-root-tls-secret"
	const defaultToken = "relay-identity-token-secret"

	if opts.RootCASecretName != "" && (opts.ClientCertSecretName != "" || opts.TokenSecretName != "") {
		// we have all the data we need: root ca and either a client cert or a token.
		// nothing to be done here
		return nil
	}
	mgmtKubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	remoteKubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.RemoteContext)
	if err != nil {
		return err
	}
	mgmtKubeSecretClient := v1.NewSecretClient(mgmtKubeClient)
	remoteKubeSecretClient := v1.NewSecretClient(remoteKubeClient)

	if opts.RootCASecretName == "" {
		opts.RootCASecretName = defaultRootCA
		if opts.RootCASecretNamespace == "" {
			opts.RootCASecretNamespace = opts.RemoteNamespace
		}
		mgmtRootCaNameNamespace := client.ObjectKey{
			Name:      defaultRootCA,
			Namespace: opts.MgmtNamespace,
		}
		// no root cert, try copy it over
		s, err := mgmtKubeSecretClient.GetSecret(ctx, mgmtRootCaNameNamespace)
		if err != nil {
			return err
		}
		copiedSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.RootCASecretName,
				Namespace: opts.RootCASecretNamespace,
			},
			Data: s.Data,
		}
		// Write it to the remote cluster. Note that we use create to make sure we don't overwrite
		// anything that already exist.
		err = remoteKubeSecretClient.CreateSecret(ctx, copiedSecret)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	if opts.ClientCertSecretName != "" {
		// if we now have a client cert, we have everything we need
		return nil
	}

	if opts.TokenSecretName == "" {
		// no token, copy it from mgmt cluster:
		opts.TokenSecretName = defaultToken
		if opts.TokenSecretNamespace == "" {
			opts.TokenSecretNamespace = opts.RemoteNamespace
		}
		mgmtTokenNameNamespace := client.ObjectKey{
			Name:      defaultToken,
			Namespace: opts.MgmtNamespace,
		}
		// no root cert, try copy it over
		s, err := mgmtKubeSecretClient.GetSecret(ctx, mgmtTokenNameNamespace)
		if err != nil {
			return err
		}
		copiedSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.TokenSecretName,
				Namespace: opts.TokenSecretNamespace,
			},
			Data: s.Data,
		}
		// write it to the remote cluster
		err = remoteKubeSecretClient.CreateSecret(ctx, copiedSecret)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}

	}
	return nil
}

func RegisterCluster(ctx context.Context, opts RegistrationOptions) error {
	chartPath, err := opts.GetChartPath(ctx, opts.AgentChartPathOverride, gloomesh.EnterpriseAgentChartUriTemplate)
	if err != nil {
		return err
	}

	values := map[string]string{
		"relay.serverAddress": opts.RelayServerAddress,
		"relay.authority":     "enterprise-networking.gloo-mesh",
		"relay.insecure":      strconv.FormatBool(opts.RelayServerInsecure),
		"relay.cluster":       opts.ClusterName,
	}

	if !opts.RelayServerInsecure {
		// read root cert from existing cluster if not provided in command line
		err = ensureCerts(ctx, opts)
		if err != nil {
			return err
		}

		if opts.RootCASecretName != "" {
			values["relay.rootTlsSecret.name"] = opts.RootCASecretName
			values["relay.rootTlsSecret.namespace"] = opts.RootCASecretNamespace
		}

		if opts.ClientCertSecretName != "" {
			values["relay.clientCertSecret.name"] = opts.RootCASecretName
			values["relay.clientCertSecret.namespace"] = opts.RootCASecretNamespace
		} else if opts.TokenSecretName != "" {
			values["relay.tokenSecret.name"] = opts.TokenSecretName
			values["relay.tokenSecret.namespace"] = opts.TokenSecretNamespace
			values["relay.tokenSecret.key"] = opts.TokenSecretKey
		}
	}

	if err := (helm.Installer{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		ChartUri:    chartPath,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: gloomesh.EnterpriseAgentReleaseName,
		ValuesFile:  opts.AgentChartValuesPath,
		Verbose:     opts.Verbose,
		Values:      values,
	}).InstallChart(ctx); err != nil {
		return err
	}

	kubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	clusterClient := v1alpha1.NewKubernetesClusterClient(kubeClient)
	return clusterClient.CreateKubernetesCluster(ctx, &v1alpha1.KubernetesCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ClusterName,
			Namespace: opts.MgmtNamespace,
		},
		Spec: v1alpha1.KubernetesClusterSpec{
			ClusterDomain: opts.ClusterDomain,
		},
	})
}

func DeregisterCluster(ctx context.Context, opts RegistrationOptions) error {
	if err := (helm.Uninstaller{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: gloomesh.EnterpriseAgentReleaseName,
		Verbose:     opts.Verbose,
	}).UninstallChart(ctx); err != nil {
		return err
	}

	kubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	clusterKey := client.ObjectKey{Name: opts.ClusterName, Namespace: opts.MgmtNamespace}
	return v1alpha1.NewKubernetesClusterClient(kubeClient).DeleteKubernetesCluster(ctx, clusterKey)
}
