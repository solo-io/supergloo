package cluster_registration_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/common/cluster-registration"
	constants2 "github.com/solo-io/service-mesh-hub/pkg/common/constants"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/installation"
	mock_csr "github.com/solo-io/service-mesh-hub/pkg/common/csr/installation/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/auth"
	mock_auth "github.com/solo-io/service-mesh-hub/pkg/common/kube/auth/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/helm"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	mock_k8s_cliendcmd "github.com/solo-io/service-mesh-hub/test/mocks/client-go/clientcmd"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
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
		mockKubeConverter           *mock_kubeconfig.MockConverter
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
		mockKubeConverter = mock_kubeconfig.NewMockConverter(ctrl)
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
			func(helmInstallerFactory helm.HelmInstallerFactory) installation.CsrAgentInstaller {
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
			GetKubernetesCluster(ctx, client.ObjectKey{Name: remoteClusterName, Namespace: container_runtime.GetWriteNamespace()}).
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
		serviceAccountRef := &smh_core_types.ResourceRef{
			Name:      remoteClusterName,
			Namespace: remoteWriteNamespace,
		}
		bearerToken := "bearer token"
		mockClusterAuthClient.EXPECT().BuildRemoteBearerToken(ctx, cfg, serviceAccountRef).Return(bearerToken, nil)
		return bearerToken
	}

	var expectInstallRemoteCSRAgent = func(useDevCsrAgentChart bool, restCfg *rest.Config) {
		mockRemoteConfig.EXPECT().ClientConfig().Return(restCfg, nil)
		mockCsrAgentInstaller.EXPECT().Install(ctx, &installation.CsrAgentInstallOptions{
			KubeConfig:           installation.KubeConfig{KubeConfig: mockRemoteConfig},
			SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
			UseDevCsrAgentChart:  useDevCsrAgentChart,
			ReleaseName:          constants2.CsrAgentReleaseName,
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
			container_runtime.GetWriteNamespace(),
			&kubeconfig.KubeConfig{
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
			GetSecret(ctx, selection.ObjectMetaToObjectKey(secret.ObjectMeta)).
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
			UpsertKubernetesClusterSpec(ctx, &smh_discovery.KubernetesCluster{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      remoteClusterName,
					Namespace: container_runtime.GetWriteNamespace(),
					Labels:    map[string]string{kube.DISCOVERED_BY: discoverySource},
				},
				Spec: smh_discovery_types.KubernetesClusterSpec{
					SecretRef: &smh_core_types.ResourceRef{
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
