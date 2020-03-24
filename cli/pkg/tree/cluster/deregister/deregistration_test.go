package deregister_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/deregister"
	mock_crd_uninstall "github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall/crd/mocks"
	helm_uninstall "github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall/helm"
	mock_helm_uninstall "github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall/helm/mocks"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/clients"
	mock_kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	mock_cli_runtime "github.com/solo-io/mesh-projects/test/mocks/cli_runtime"
	"helm.sh/helm/v3/pkg/action"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
)

var _ = Describe("Cluster Deregistration", func() {
	var (
		ctrl             *gomock.Controller
		ctx              context.Context
		remoteRestConfig = &rest.Config{
			Host: "remote-cluster.com",
		}
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can deregister a cluster", func() {
		secretsClient := mock_kubernetes_core.NewMockSecretsClient(ctrl)
		helmUninstaller := mock_helm_uninstall.NewMockUninstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)
		kubeConfigSecretRef := &core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.DefaultWriteNamespace,
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &discovery_v1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		kubeConfigSecret := &v1.Secret{
			ObjectMeta: clients.ResourceRefToObjectMeta(kubeConfigSecretRef),
		}

		secretsClient.EXPECT().
			Get(ctx, "kube-config-secret", env.DefaultWriteNamespace).
			Return(kubeConfigSecret, nil)
		var secretToConfigConverter kubeconfig.SecretToConfigConverter = func(secret *v1.Secret) (clusterName string, config *kubeconfig.Config, err error) {
			Expect(secret).To(Equal(kubeConfigSecret))
			return remoteClusterName, &kubeconfig.Config{
				RestConfig: remoteRestConfig,
			}, nil
		}
		helmUninstaller.EXPECT().
			Run(cliconstants.CsrAgentReleaseName).
			Return(nil, nil)
		crdRemover.EXPECT().
			RemoveZephyrCrds(remoteClusterName, remoteRestConfig).
			Return(true, nil)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			secretsClient,
			secretToConfigConverter,
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Expect(cfg).To(Equal(remoteRestConfig))
				return restClientGetter
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Expect(namespace).To(Equal(remoteWriteNamespace))

				return helmUninstaller, nil
			},
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with the appropriate error if it can't find the kube config secret", func() {
		testErr := eris.New("test-err")
		secretsClient := mock_kubernetes_core.NewMockSecretsClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		kubeConfigSecretRef := &core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.DefaultWriteNamespace,
		}
		remoteWriteNamespace := "remote-write-namespace"
		remoteClusterName := "remote-cluster-name"
		clusterToDeregister := &discovery_v1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}

		secretsClient.EXPECT().
			Get(ctx, "kube-config-secret", env.DefaultWriteNamespace).
			Return(nil, testErr)
		var secretToConfigConverter kubeconfig.SecretToConfigConverter = func(secret *v1.Secret) (clusterName string, config *kubeconfig.Config, err error) {
			Fail("Should not be called in this test")
			return "", nil, nil
		}

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			secretsClient,
			secretToConfigConverter,
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Fail("Should have called the rest client getter factory")
				return nil
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Fail("Should not have called the helm uninstaller factory")
				return nil, nil
			},
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(deregister.FailedToFindKubeConfigSecret(testErr, remoteClusterName)))
	})

	It("responds with the appropriate error if CSR uninstallation fails", func() {
		testErr := eris.New("test-err")
		secretsClient := mock_kubernetes_core.NewMockSecretsClient(ctrl)
		helmUninstaller := mock_helm_uninstall.NewMockUninstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)
		kubeConfigSecretRef := &core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.DefaultWriteNamespace,
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &discovery_v1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		kubeConfigSecret := &v1.Secret{
			ObjectMeta: clients.ResourceRefToObjectMeta(kubeConfigSecretRef),
		}

		secretsClient.EXPECT().
			Get(ctx, "kube-config-secret", env.DefaultWriteNamespace).
			Return(kubeConfigSecret, nil)
		var secretToConfigConverter kubeconfig.SecretToConfigConverter = func(secret *v1.Secret) (clusterName string, config *kubeconfig.Config, err error) {
			Expect(secret).To(Equal(kubeConfigSecret))
			return remoteClusterName, &kubeconfig.Config{
				RestConfig: remoteRestConfig,
			}, nil
		}
		helmUninstaller.EXPECT().
			Run(cliconstants.CsrAgentReleaseName).
			Return(nil, testErr)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			secretsClient,
			secretToConfigConverter,
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Expect(cfg).To(Equal(remoteRestConfig))
				return restClientGetter
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Expect(namespace).To(Equal(remoteWriteNamespace))

				return helmUninstaller, nil
			},
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(deregister.FailedToUninstallCsrAgent(testErr, remoteClusterName)))
	})

	It("responds with the appropriate error if CRD removal fails", func() {
		testErr := eris.New("test-err")
		secretsClient := mock_kubernetes_core.NewMockSecretsClient(ctrl)
		helmUninstaller := mock_helm_uninstall.NewMockUninstaller(ctrl)
		restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		kubeConfigSecretRef := &core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.DefaultWriteNamespace,
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &discovery_v1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		kubeConfigSecret := &v1.Secret{
			ObjectMeta: clients.ResourceRefToObjectMeta(kubeConfigSecretRef),
		}

		secretsClient.EXPECT().
			Get(ctx, "kube-config-secret", env.DefaultWriteNamespace).
			Return(kubeConfigSecret, nil)
		var secretToConfigConverter kubeconfig.SecretToConfigConverter = func(secret *v1.Secret) (clusterName string, config *kubeconfig.Config, err error) {
			Expect(secret).To(Equal(kubeConfigSecret))
			return remoteClusterName, &kubeconfig.Config{
				RestConfig: remoteRestConfig,
			}, nil
		}
		helmUninstaller.EXPECT().
			Run(cliconstants.CsrAgentReleaseName).
			Return(nil, nil)
		crdRemover.EXPECT().
			RemoveZephyrCrds(remoteClusterName, remoteRestConfig).
			Return(false, testErr)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			secretsClient,
			secretToConfigConverter,
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Expect(cfg).To(Equal(remoteRestConfig))
				return restClientGetter
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Expect(namespace).To(Equal(remoteWriteNamespace))

				return helmUninstaller, nil
			},
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(deregister.FailedToRemoveCrds(testErr, remoteClusterName)))
	})
})
