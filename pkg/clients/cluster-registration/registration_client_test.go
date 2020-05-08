package cluster_registration_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	mock_kube "github.com/solo-io/service-mesh-hub/cli/pkg/common/kube/mocks"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/auth/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/clients/cluster-registration"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/solo-io/service-mesh-hub/pkg/installers/csr"
	mock_csr "github.com/solo-io/service-mesh-hub/pkg/installers/csr/mocks"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mock_k8s_cliendcmd "github.com/solo-io/service-mesh-hub/test/mocks/client-go/clientcmd"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ClusterRegistrationClient", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockSecretClient            *mock_kubernetes_core.MockSecretClient
		mockKubernetesClusterClient *mock_core.MockKubernetesClusterClient
		mockNamespaceClient         *mock_kubernetes_core.MockNamespaceClient
		mockKubeConverter           *mock_kube.MockConverter
		mockCsrAgentInstaller       *mock_csr.MockCsrAgentInstaller
		mockClusterAuthClient       *mock_auth.MockClusterAuthorization
		clusterRegistrationClient   cluster_registration.ClusterRegistrationClient
		mockRemoteConfig            *mock_k8s_cliendcmd.MockClientConfig
		remoteClusterName           = "remote-cluster-name"
		remoteWriteNamespace        = "remote-write-namespace"
		remoteContextName           = "remote-context-name"
		discoverySource             = "discovery-source"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockSecretClient = mock_kubernetes_core.NewMockSecretClient(ctrl)
		mockKubernetesClusterClient = mock_core.NewMockKubernetesClusterClient(ctrl)
		mockNamespaceClient = mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		mockKubeConverter = mock_kube.NewMockConverter(ctrl)
		mockCsrAgentInstaller = mock_csr.NewMockCsrAgentInstaller(ctrl)
		mockRemoteConfig = mock_k8s_cliendcmd.NewMockClientConfig(ctrl)
		mockClusterAuthClient = mock_auth.NewMockClusterAuthorization(ctrl)
		clusterRegistrationClient = cluster_registration.NewClusterRegistrationClient(
			mockSecretClient,
			mockKubernetesClusterClient,
			func(cfg *rest.Config) (k8s_core.NamespaceClient, error) {
				return mockNamespaceClient, nil
			},
			mockKubeConverter,
			func(helmInstallerFactory factories.HelmerInstallerFactory) csr.CsrAgentInstaller {
				return mockCsrAgentInstaller
			},
			func(remoteAuthConfig *rest.Config) (auth.ClusterAuthorization, error) {
				return mockClusterAuthClient, nil
			})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectClusterNotExists = func() {
		mockKubernetesClusterClient.
			EXPECT().
			GetKubernetesCluster(ctx, client.ObjectKey{Name: remoteClusterName, Namespace: env.GetWriteNamespace()}).
			Return(nil, errors.NewNotFound(controllerruntime.GroupResource{}, "test-resource"))
	}

	var expectCreateRemoteNamespace = func() {
		mockNamespaceClient.
			EXPECT().
			GetNamespace(ctx, client.ObjectKey{Name: remoteWriteNamespace}).
			Return(nil, errors.NewNotFound(controllerruntime.GroupResource{}, "test-resource"))
		mockNamespaceClient.
			EXPECT().
			CreateNamespace(ctx, &k8s_core_types.Namespace{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: remoteWriteNamespace,
				},
			}).
			Return(nil)
	}

	var expectGenServiceAccountBearerToken = func(cfg *rest.Config) string {
		serviceAccountRef := &zephyr_core_types.ResourceRef{
			Name:      remoteClusterName,
			Namespace: remoteWriteNamespace,
		}
		bearerToken := "bearer token"
		mockClusterAuthClient.EXPECT().BuildRemoteBearerToken(ctx, cfg, serviceAccountRef).Return(bearerToken, nil)
		return bearerToken
	}

	var expectInstallRemoteCSRAgent = func(useDevCsrAgentChart bool, restCfg *rest.Config) {
		mockRemoteConfig.EXPECT().ClientConfig().Return(restCfg, nil)
		mockCsrAgentInstaller.EXPECT().Install(ctx, &csr.CsrAgentInstallOptions{
			KubeConfig:           mockRemoteConfig,
			ClusterName:          remoteClusterName,
			SmhInstallNamespace:  env.GetWriteNamespace(),
			UseDevCsrAgentChart:  useDevCsrAgentChart,
			ReleaseName:          cliconstants.CsrAgentReleaseName,
			RemoteWriteNamespace: remoteWriteNamespace,
		}).Return(nil)
	}

	var expectWriteKubeConfigToMaster = func(bearerToken, localClusterDomainOverride string) *k8s_core_types.Secret {
		apiConfig := api.Config{
			Contexts: map[string]*api.Context{
				remoteContextName: {
					LocationOfOrigin: "context-location-of-origin",
					Cluster:          "cluster",
					AuthInfo:         "",
					Namespace:        "context-namespace",
					Extensions:       map[string]runtime.Object{},
				},
			},
			Clusters: map[string]*api.Cluster{
				"cluster": {
					LocationOfOrigin:         "location-of-origin",
					Server:                   "",
					InsecureSkipTLSVerify:    false,
					CertificateAuthority:     "",
					CertificateAuthorityData: nil,
					Extensions:               nil,
				},
			},
		}
		mockRemoteConfig.EXPECT().RawConfig().Return(apiConfig, nil)
		secret := &k8s_core_types.Secret{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "name", Namespace: "namespace"}}
		mockKubeConverter.EXPECT().ConfigToSecret(
			remoteClusterName,
			env.GetWriteNamespace(),
			&kube.KubeConfig{
				Config: api.Config{
					Kind:        "Secret",
					APIVersion:  "kubernetes_core",
					Preferences: api.Preferences{},
					Clusters: map[string]*api.Cluster{
						remoteClusterName: apiConfig.Clusters["cluster"],
					},
					AuthInfos: map[string]*api.AuthInfo{
						remoteClusterName: {
							Token: bearerToken,
						},
					},
					Contexts: map[string]*api.Context{
						remoteClusterName: {
							LocationOfOrigin: apiConfig.Contexts[remoteContextName].LocationOfOrigin,
							Cluster:          remoteClusterName,
							AuthInfo:         remoteClusterName,
							Namespace:        apiConfig.Contexts[remoteContextName].Namespace,
							Extensions:       apiConfig.Contexts[remoteContextName].Extensions,
						},
					},
					CurrentContext: remoteClusterName,
				},
				Cluster: remoteClusterName,
			}).Return(secret, nil)
		mockSecretClient.
			EXPECT().
			GetSecret(ctx, clients.ObjectMetaToObjectKey(secret.ObjectMeta)).
			Return(nil, errors.NewNotFound(controllerruntime.GroupResource{}, "test-resource"))
		mockSecretClient.
			EXPECT().
			CreateSecret(ctx, secret).
			Return(nil)
		return secret
	}

	var expectWriteKubeClusterToMaster = func(secret *k8s_core_types.Secret) {
		mockKubernetesClusterClient.
			EXPECT().
			UpsertKubernetesClusterSpec(ctx, &zephyr_discovery.KubernetesCluster{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      remoteClusterName,
					Namespace: env.GetWriteNamespace(),
					Labels:    map[string]string{constants.DISCOVERED_BY: discoverySource},
				},
				Spec: zephyr_discovery_types.KubernetesClusterSpec{
					SecretRef: &zephyr_core_types.ResourceRef{
						Name:      secret.GetName(),
						Namespace: secret.GetNamespace(),
					},
					WriteNamespace: remoteWriteNamespace,
				},
			}).Return(nil)
	}

	It("should register a cluster", func() {
		useDevCsrAgentChart := false
		localClusterDomainOverride := ""
		restConfig := &rest.Config{}
		mockRemoteConfig.EXPECT().ClientConfig().Return(restConfig, nil)
		expectClusterNotExists()
		expectCreateRemoteNamespace()
		bearerToken := expectGenServiceAccountBearerToken(restConfig)
		expectInstallRemoteCSRAgent(useDevCsrAgentChart, restConfig)
		secret := expectWriteKubeConfigToMaster(bearerToken, localClusterDomainOverride)
		expectWriteKubeClusterToMaster(secret)

		err := clusterRegistrationClient.Register(
			ctx,
			mockRemoteConfig,
			remoteClusterName,
			remoteWriteNamespace,
			remoteContextName,
			discoverySource,
			cluster_registration.ClusterRegisterOpts{
				Overwrite:                  false,
				UseDevCsrAgentChart:        useDevCsrAgentChart,
				LocalClusterDomainOverride: localClusterDomainOverride,
			},
		)
		Expect(err).To(BeNil())
	})

	It("should return error if context not found in kubeconfig", func() {
		useDevCsrAgentChart := false
		localClusterDomainOverride := ""
		restConfig := &rest.Config{}
		mockRemoteConfig.EXPECT().ClientConfig().Return(restConfig, nil)
		expectClusterNotExists()
		expectCreateRemoteNamespace()
		expectGenServiceAccountBearerToken(restConfig)
		expectInstallRemoteCSRAgent(useDevCsrAgentChart, restConfig)

		apiConfig := api.Config{
			Contexts: map[string]*api.Context{
				"some other name": {},
			},
		}
		mockRemoteConfig.EXPECT().RawConfig().Return(apiConfig, nil)

		err := clusterRegistrationClient.Register(
			ctx,
			mockRemoteConfig,
			remoteClusterName,
			remoteWriteNamespace,
			remoteContextName,
			discoverySource,
			cluster_registration.ClusterRegisterOpts{
				Overwrite:                  false,
				UseDevCsrAgentChart:        useDevCsrAgentChart,
				LocalClusterDomainOverride: localClusterDomainOverride,
			},
		)
		Expect(err).To(testutils.HaveInErrorChain(cluster_registration.ContextNotFound(remoteContextName)))
	})

	It("should return error if context not found in kubeconfig", func() {
		useDevCsrAgentChart := false
		localClusterDomainOverride := ""
		restConfig := &rest.Config{}
		mockRemoteConfig.EXPECT().ClientConfig().Return(restConfig, nil)
		expectClusterNotExists()
		expectCreateRemoteNamespace()
		expectGenServiceAccountBearerToken(restConfig)
		expectInstallRemoteCSRAgent(useDevCsrAgentChart, restConfig)

		nonExtantClusterName := "not-found-cluster"
		apiConfig := api.Config{
			Contexts: map[string]*api.Context{
				remoteContextName: {
					LocationOfOrigin: "context-location-of-origin",
					Cluster:          nonExtantClusterName,
				},
			},
			Clusters: map[string]*api.Cluster{
				"cluster": {
					LocationOfOrigin:         "location-of-origin",
					Server:                   "",
					InsecureSkipTLSVerify:    false,
					CertificateAuthority:     "",
					CertificateAuthorityData: nil,
					Extensions:               nil,
				},
			},
		}
		mockRemoteConfig.EXPECT().RawConfig().Return(apiConfig, nil)

		err := clusterRegistrationClient.Register(
			ctx,
			mockRemoteConfig,
			remoteClusterName,
			remoteWriteNamespace,
			remoteContextName,
			discoverySource,
			cluster_registration.ClusterRegisterOpts{
				Overwrite:                  false,
				UseDevCsrAgentChart:        useDevCsrAgentChart,
				LocalClusterDomainOverride: localClusterDomainOverride,
			},
		)
		Expect(err).To(testutils.HaveInErrorChain(cluster_registration.ClusterNotFound(nonExtantClusterName)))
	})
})
