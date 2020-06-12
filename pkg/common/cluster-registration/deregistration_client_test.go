package cluster_registration_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	mock_k8s_core_clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_smh_discovery_clients "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_security_scheme "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	cluster_registration2 "github.com/solo-io/service-mesh-hub/pkg/common/cluster-registration"
	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	cert_secrets2 "github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen/secrets"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/installation"
	mock_installation "github.com/solo-io/service-mesh-hub/pkg/common/csr/installation/mocks"
	auth2 "github.com/solo-io/service-mesh-hub/pkg/common/kube/auth"
	mock_crd_uninstall "github.com/solo-io/service-mesh-hub/pkg/common/kube/crd/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/helm"
	kubeconfig2 "github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	mock_kubeconfig2 "github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig/mocks"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	v12 "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Cluster Deregistration", func() {
	var (
		ctrl             *gomock.Controller
		ctx              context.Context
		remoteRestConfig = &rest.Config{
			Host: "remote-cluster.com",
		}
		remoteClientConfig = &clientcmd.DirectClientConfig{}
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can deregister a cluster", func() {
		mockCsrAgentInstaller := mock_installation.NewMockCsrAgentInstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		configLookup := mock_kubeconfig2.NewMockKubeConfigLookup(ctrl)
		kubeClusterClient := mock_smh_discovery_clients.NewMockKubernetesClusterClient(ctrl)
		localSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		remoteSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		dynamicClientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
		remoteServiceAccountClient := mock_k8s_core_clients.NewMockServiceAccountClient(ctrl)

		kubeConfigSecretRef := &smh_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		kubeConfigSecret := &v12.Secret{
			ObjectMeta: selection.ResourceRefToObjectMeta(kubeConfigSecretRef),
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		intermediateCertSecret := &v12.Secret{
			Type: cert_secrets2.IntermediateCertSecretType,
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "intermediate-cert",
				Namespace: remoteWriteNamespace,
			},
		}

		kubeRestConfig := &kubeconfig2.ConvertedConfigs{
			RestConfig:   remoteRestConfig,
			ClientConfig: remoteClientConfig,
		}
		mockCsrAgentInstaller.EXPECT().Uninstall(&installation.CsrAgentUninstallOptions{
			KubeConfig:       installation.KubeConfig{KubeConfig: remoteClientConfig},
			ReleaseName:      constants.CsrAgentReleaseName,
			ReleaseNamespace: clusterToDeregister.Spec.GetWriteNamespace(),
		}).Return(nil)
		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(kubeRestConfig, nil)
		dynamicClientGetter.EXPECT().
			GetClientForCluster(ctx, remoteClusterName).
			Return(nil, nil)
		remoteSecretClient.EXPECT().
			ListSecret(ctx, client.InNamespace(remoteWriteNamespace)).
			Return(&v12.SecretList{
				Items: []v12.Secret{*intermediateCertSecret},
			}, nil)
		remoteSecretClient.EXPECT().
			DeleteSecret(ctx, selection.ObjectMetaToObjectKey(intermediateCertSecret.ObjectMeta)).
			Return(nil)
		localSecretClient.EXPECT().
			GetSecret(ctx, selection.ResourceRefToObjectKey(clusterToDeregister.Spec.GetSecretRef())).
			Return(kubeConfigSecret, nil)
		localSecretClient.EXPECT().
			DeleteSecret(ctx, selection.ObjectMetaToObjectKey(kubeConfigSecret.ObjectMeta)).
			Return(nil)
		kubeClusterClient.EXPECT().
			DeleteKubernetesCluster(ctx, selection.ObjectMetaToObjectKey(clusterToDeregister.ObjectMeta)).
			Return(nil)
		remoteServiceAccountClient.EXPECT().
			DeleteAllOfServiceAccount(
				ctx,
				client.InNamespace(remoteWriteNamespace),
				client.MatchingLabels{
					constants.ManagedByLabel:         constants.ServiceMeshHubApplicationName,
					auth2.RegistrationServiceAccount: auth2.RegistrationServiceAccountValue,
				},
			).
			Return(nil)
		crdRemover.EXPECT().
			RemoveCrdGroup(ctx, clusterToDeregister.GetName(), kubeRestConfig.RestConfig, smh_security_scheme.SchemeGroupVersion).
			Return(false, nil)

		clusterDeregistrationClient := cluster_registration2.NewClusterDeregistrationClient(
			crdRemover,
			func(helmInstallerFactory helm.HelmInstallerFactory) installation.CsrAgentInstaller {
				return mockCsrAgentInstaller
			},
			configLookup,
			kubeClusterClient,
			localSecretClient,
			func(_ client.Client) v1.SecretClient {
				return remoteSecretClient
			},
			dynamicClientGetter,
			func(_ client.Client) v1.ServiceAccountClient {
				return remoteServiceAccountClient
			},
		)

		err := clusterDeregistrationClient.Deregister(ctx, clusterToDeregister)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with the appropriate error if the config lookup fails", func() {
		testErr := eris.New("test-err")
		mockCsrAgentInstaller := mock_installation.NewMockCsrAgentInstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		configLookup := mock_kubeconfig2.NewMockKubeConfigLookup(ctrl)
		kubeClusterClient := mock_smh_discovery_clients.NewMockKubernetesClusterClient(ctrl)
		localSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		dynamicClientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
		kubeConfigSecretRef := &smh_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		remoteWriteNamespace := "remote-write-namespace"
		remoteClusterName := "remote-cluster-name"
		clusterToDeregister := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}

		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(nil, testErr)

		clusterDeregistrationClient := cluster_registration2.NewClusterDeregistrationClient(
			crdRemover,
			func(helmInstallerFactory helm.HelmInstallerFactory) installation.CsrAgentInstaller {
				return mockCsrAgentInstaller
			},
			configLookup,
			kubeClusterClient,
			localSecretClient,
			func(_ client.Client) v1.SecretClient {
				return nil
			},
			dynamicClientGetter,
			func(_ client.Client) v1.ServiceAccountClient {
				return nil
			},
		)

		err := clusterDeregistrationClient.Deregister(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(cluster_registration2.FailedToFindClusterCredentials(testErr, remoteClusterName)))
	})

	It("responds with the appropriate error if CSR uninstallation fails", func() {
		testErr := eris.New("test-err")
		mockCsrAgentInstaller := mock_installation.NewMockCsrAgentInstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		configLookup := mock_kubeconfig2.NewMockKubeConfigLookup(ctrl)
		kubeClusterClient := mock_smh_discovery_clients.NewMockKubernetesClusterClient(ctrl)
		localSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		remoteSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		dynamicClientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
		remoteServiceAccountClient := mock_k8s_core_clients.NewMockServiceAccountClient(ctrl)
		kubeConfigSecretRef := &smh_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(&kubeconfig2.ConvertedConfigs{
				RestConfig:   remoteRestConfig,
				ClientConfig: remoteClientConfig,
			}, nil)
		mockCsrAgentInstaller.EXPECT().Uninstall(&installation.CsrAgentUninstallOptions{
			KubeConfig:       installation.KubeConfig{KubeConfig: remoteClientConfig},
			ReleaseName:      constants.CsrAgentReleaseName,
			ReleaseNamespace: clusterToDeregister.Spec.GetWriteNamespace(),
		}).Return(testErr)

		clusterDeregistrationClient := cluster_registration2.NewClusterDeregistrationClient(
			crdRemover,
			func(helmInstallerFactory helm.HelmInstallerFactory) installation.CsrAgentInstaller {
				return mockCsrAgentInstaller
			},
			configLookup,
			kubeClusterClient,
			localSecretClient,
			func(_ client.Client) v1.SecretClient {
				return remoteSecretClient
			},
			dynamicClientGetter,
			func(_ client.Client) v1.ServiceAccountClient {
				return remoteServiceAccountClient
			},
		)

		err := clusterDeregistrationClient.Deregister(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(cluster_registration2.FailedToUninstallCsrAgent(testErr, remoteClusterName)))
	})

	It("responds with the appropriate error if CRD removal fails", func() {
		testErr := eris.New("test-err")
		mockCsrAgentInstaller := mock_installation.NewMockCsrAgentInstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		configLookup := mock_kubeconfig2.NewMockKubeConfigLookup(ctrl)
		kubeClusterClient := mock_smh_discovery_clients.NewMockKubernetesClusterClient(ctrl)
		localSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		remoteSecretClient := mock_k8s_core_clients.NewMockSecretClient(ctrl)
		dynamicClientGetter := mock_multicluster.NewMockDynamicClientGetter(ctrl)
		remoteServiceAccountClient := mock_k8s_core_clients.NewMockServiceAccountClient(ctrl)
		kubeConfigSecretRef := &smh_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		kubeConfigSecret := &v12.Secret{
			ObjectMeta: selection.ResourceRefToObjectMeta(kubeConfigSecretRef),
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}

		intermediateCertSecret := &v12.Secret{
			Type: cert_secrets2.IntermediateCertSecretType,
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "intermediate-cert",
				Namespace: remoteWriteNamespace,
			},
		}
		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(&kubeconfig2.ConvertedConfigs{
				RestConfig:   remoteRestConfig,
				ClientConfig: remoteClientConfig,
			}, nil)
		mockCsrAgentInstaller.EXPECT().Uninstall(&installation.CsrAgentUninstallOptions{
			KubeConfig:       installation.KubeConfig{KubeConfig: remoteClientConfig},
			ReleaseName:      constants.CsrAgentReleaseName,
			ReleaseNamespace: clusterToDeregister.Spec.GetWriteNamespace(),
		}).Return(nil)
		crdRemover.EXPECT().
			RemoveCrdGroup(ctx, clusterToDeregister.GetName(), remoteRestConfig, smh_security_scheme.SchemeGroupVersion).
			Return(false, testErr)
		dynamicClientGetter.EXPECT().
			GetClientForCluster(ctx, remoteClusterName).
			Return(nil, nil)
		remoteSecretClient.EXPECT().
			ListSecret(ctx, client.InNamespace(remoteWriteNamespace)).
			Return(&v12.SecretList{
				Items: []v12.Secret{*intermediateCertSecret},
			}, nil)
		remoteSecretClient.EXPECT().
			DeleteSecret(ctx, selection.ObjectMetaToObjectKey(intermediateCertSecret.ObjectMeta)).
			Return(nil)
		localSecretClient.EXPECT().
			GetSecret(ctx, selection.ResourceRefToObjectKey(clusterToDeregister.Spec.GetSecretRef())).
			Return(kubeConfigSecret, nil)
		localSecretClient.EXPECT().
			DeleteSecret(ctx, selection.ObjectMetaToObjectKey(kubeConfigSecret.ObjectMeta)).
			Return(nil)
		kubeClusterClient.EXPECT().
			DeleteKubernetesCluster(ctx, selection.ObjectMetaToObjectKey(clusterToDeregister.ObjectMeta)).
			Return(nil)

		clusterDeregistrationClient := cluster_registration2.NewClusterDeregistrationClient(
			crdRemover,
			func(helmInstallerFactory helm.HelmInstallerFactory) installation.CsrAgentInstaller {
				return mockCsrAgentInstaller
			},
			configLookup,
			kubeClusterClient,
			localSecretClient,
			func(_ client.Client) v1.SecretClient {
				return remoteSecretClient
			},
			dynamicClientGetter,
			func(_ client.Client) v1.ServiceAccountClient {
				return remoteServiceAccountClient
			},
		)

		err := clusterDeregistrationClient.Deregister(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(cluster_registration2.FailedToRemoveCrds(testErr, remoteClusterName)))
	})
})
