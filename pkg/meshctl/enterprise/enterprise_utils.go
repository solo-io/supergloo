package enterprise

import (
	"context"
	"fmt"
	"strconv"
	"time"

	. "github.com/logrusorgru/aurora/v3"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
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

	ReleaseName string
}

func ensureCerts(ctx context.Context, opts *RegistrationOptions) (bool, error) {
	// if secure, and no user data was given, attempt to deduce the required parameters.
	const defaultRootCA = "relay-root-tls-secret"
	const defaultToken = "relay-identity-token-secret"
	const defaultTokenSecretKey = "token"

	createdBootstrapToken := false

	if opts.RootCASecretName != "" && (opts.ClientCertSecretName != "" || opts.TokenSecretName != "") {
		// we have all the data we need: root ca and either a client cert or a token.
		// nothing to be done here
		return createdBootstrapToken, nil
	}
	mgmtKubeConfigPath := opts.KubeConfigPath
	// override if provided
	if opts.MgmtKubeConfigPath != "" {
		mgmtKubeConfigPath = opts.MgmtKubeConfigPath
	}
	mgmtKubeClient, err := utils.BuildClient(mgmtKubeConfigPath, opts.MgmtContext)
	if err != nil {
		return createdBootstrapToken, err
	}
	remoteKubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.RemoteContext)
	if err != nil {
		return createdBootstrapToken, err
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

		if err = utils.EnsureNamespace(ctx, remoteKubeClient, opts.RemoteNamespace); err != nil {
			return createdBootstrapToken, eris.Wrapf(err, "creating namespace")
		}
		// no root cert, try copy it over
		logrus.Info("📃 Copying root CA ", Bold(fmt.Sprintf("%s.%s", mgmtRootCaNameNamespace.Name, mgmtRootCaNameNamespace.Namespace)), " to remote cluster from management cluster")

		s, err := mgmtKubeSecretClient.GetSecret(ctx, mgmtRootCaNameNamespace)
		if err != nil {
			return createdBootstrapToken, err
		}
		copiedSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.RootCASecretName,
				Namespace: opts.RootCASecretNamespace,
			},
			Data: map[string][]byte{
				"ca.crt": s.Data["ca.crt"],
			},
		}
		// Write it to the remote cluster. Note that we use create to make sure we don't overwrite
		// anything that already exist.
		err = remoteKubeSecretClient.CreateSecret(ctx, copiedSecret)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return createdBootstrapToken, err
		}
	}

	if opts.ClientCertSecretName != "" {
		// if we now have a client cert, we have everything we need
		return createdBootstrapToken, nil
	}

	if opts.TokenSecretName == "" {
		// no token, copy it from mgmt cluster:
		opts.TokenSecretName = defaultToken
		if opts.TokenSecretNamespace == "" {
			opts.TokenSecretNamespace = opts.RemoteNamespace
		}
		if opts.TokenSecretKey == "" {
			opts.TokenSecretKey = defaultTokenSecretKey
		}
		mgmtTokenNameNamespace := client.ObjectKey{
			Name:      defaultToken,
			Namespace: opts.MgmtNamespace,
		}
		if err = utils.EnsureNamespace(ctx, remoteKubeClient, opts.RemoteNamespace); err != nil {
			return createdBootstrapToken, eris.Wrapf(err, "creating namespace")
		}
		logrus.Info("📃 Copying bootstrap token ", Bold(fmt.Sprintf("%s.%s", opts.TokenSecretName, opts.TokenSecretNamespace)), " to remote cluster from management cluster")
		// no root cert, try copy it over
		s, err := mgmtKubeSecretClient.GetSecret(ctx, mgmtTokenNameNamespace)
		if err != nil {
			return createdBootstrapToken, err
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
			// A write error occurred.
			return createdBootstrapToken, err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			// The user either provisioned their own token secret, or
			// we're on the management cluster and using the server's
			// token secret.
			createdBootstrapToken = false
		} else {
			// Successfully created the bootstrap token on a remote cluster.
			createdBootstrapToken = true
		}
	}
	return createdBootstrapToken, nil
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
		"global.insecure":     strconv.FormatBool(opts.RelayServerInsecure),
		"relay.cluster":       opts.ClusterName,
	}
	bootstrapTokenCreated := false
	if !opts.RelayServerInsecure {
		// read root cert from existing cluster if not provided in command line
		bootstrapTokenCreated, err = ensureCerts(ctx, &opts)
		if err != nil {
			return err
		}

		// we should have root tls name by here
		values["relay.rootTlsSecret.name"] = opts.RootCASecretName
		values["relay.rootTlsSecret.namespace"] = opts.RootCASecretNamespace

		// relay needs a client cert provided, even if it doesnt exist, so it can write to it.
		if opts.ClientCertSecretName == "" {
			opts.ClientCertSecretName = "relay-client-tls-secret"
		}
		if opts.ClientCertSecretNamespace == "" {
			opts.ClientCertSecretNamespace = opts.RemoteNamespace
		}
		values["relay.clientCertSecret.name"] = opts.ClientCertSecretName
		values["relay.clientCertSecret.namespace"] = opts.ClientCertSecretNamespace

		// only copy token secret if we have one
		if opts.TokenSecretName != "" {
			values["relay.tokenSecret.name"] = opts.TokenSecretName
			values["relay.tokenSecret.namespace"] = opts.TokenSecretNamespace
			values["relay.tokenSecret.key"] = opts.TokenSecretKey
		}
	}
	logrus.Info("💻 Installing relay agent in the remote cluster")

	releaseName := gloomesh.EnterpriseAgentReleaseName
	if opts.ReleaseName != "" {
		releaseName = opts.ReleaseName
	}

	if err := (helm.Installer{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		ChartUri:    chartPath,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: releaseName,
		ValuesFile:  opts.AgentChartValuesPath,
		Verbose:     opts.Verbose,
		Values:      values,
	}).InstallChart(ctx); err != nil {
		return err
	}

	kubeConfigPath := opts.MgmtKubeConfigPath
	if kubeConfigPath == "" {
		kubeConfigPath = opts.KubeConfigPath
	}

	mgmtKubeClient, err := utils.BuildClient(kubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	mgmtClusterClient := v1alpha1.NewKubernetesClusterClient(mgmtKubeClient)

	logrus.Info("📃 Creating ", Bold(opts.ClusterName+" KubernetesCluster CRD"), " in management cluster")

	err = mgmtClusterClient.CreateKubernetesCluster(ctx, &v1alpha1.KubernetesCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ClusterName,
			Namespace: opts.MgmtNamespace,
		},
		Spec: v1alpha1.KubernetesClusterSpec{
			ClusterDomain: opts.ClusterDomain,
		},
	})
	if err != nil {
		return err
	}

	if !opts.RelayServerInsecure {
		logrus.Info("⌚ Waiting for relay agent to have a client certificate")

		remoteKubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.RemoteContext)
		if err != nil {
			return err
		}
		remoteKubeSecretClient := v1.NewSecretClient(remoteKubeClient)

		err = waitForClientCert(ctx, remoteKubeSecretClient, opts)
		if err != nil {
			return err
		}
		if bootstrapTokenCreated {
			// Delete the bootstrap token from the registered cluster
			// if it was created by this command invocation.
			logrus.Info("🗑 Removing bootstrap token")
			key := client.ObjectKey{
				Name:      opts.TokenSecretName,
				Namespace: opts.TokenSecretNamespace,
			}
			err = remoteKubeSecretClient.DeleteSecret(ctx, key)
			if err != nil {
				return err
			}
		}
	}

	logrus.Info("✅ Done registering cluster!")
	return nil
}

func waitForClientCert(ctx context.Context, remoteKubeSecretClient v1.SecretClient, opts RegistrationOptions) error {

	clientCert := client.ObjectKey{
		Name:      opts.ClientCertSecretName,
		Namespace: opts.ClientCertSecretNamespace,
	}

	timeout := time.After(2 * time.Minute)

	for {
		_, err := remoteKubeSecretClient.GetSecret(ctx, clientCert)
		if !apierrors.IsNotFound(err) {
			return err
		}
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return eris.Errorf("timed out waiting for client cert")
		case <-time.After(5 * time.Second):
			logrus.Info("\t Checking...")
		}
	}
	logrus.Info("📃 Client certificate found in remote cluster")
	return nil
}

func DeregisterCluster(ctx context.Context, opts RegistrationOptions) error {
	releaseName := gloomesh.EnterpriseAgentReleaseName
	if opts.ReleaseName != "" {
		releaseName = opts.ReleaseName
	}
	if err := (helm.Uninstaller{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: releaseName,
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
