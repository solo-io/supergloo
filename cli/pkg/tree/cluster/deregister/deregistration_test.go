package deregister_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/deregister"
	mock_config_lookup "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/config_lookup/mocks"
	mock_crd_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/crd/mocks"
	helm_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/helm"
	mock_helm_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/helm/mocks"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_cli_runtime "github.com/solo-io/service-mesh-hub/test/mocks/cli_runtime"
	"helm.sh/helm/v3/pkg/action"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		helmUninstaller := mock_helm_uninstall.NewMockUninstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)
		configLookup := mock_config_lookup.NewMockKubeConfigLookup(ctrl)
		kubeConfigSecretRef := &zephyr_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.GetWriteNamespace(),
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &zephyr_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.GetWriteNamespace(),
			},
			Spec: zephyr_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}

		helmUninstaller.EXPECT().
			Run(cliconstants.CsrAgentReleaseName).
			Return(nil, nil)
		crdRemover.EXPECT().
			RemoveZephyrCrds(ctx, remoteClusterName, remoteRestConfig).
			Return(true, nil)
		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(&kube.ConvertedConfigs{RestConfig: remoteRestConfig}, nil)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Expect(cfg).To(Equal(remoteRestConfig))
				return restClientGetter
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Expect(namespace).To(Equal(remoteWriteNamespace))

				return helmUninstaller, nil
			},
			configLookup,
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with the appropriate error if the config lookup fails", func() {
		testErr := eris.New("test-err")
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		configLookup := mock_config_lookup.NewMockKubeConfigLookup(ctrl)
		kubeConfigSecretRef := &zephyr_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.GetWriteNamespace(),
		}
		remoteWriteNamespace := "remote-write-namespace"
		remoteClusterName := "remote-cluster-name"
		clusterToDeregister := &zephyr_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.GetWriteNamespace(),
			},
			Spec: zephyr_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}

		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(nil, testErr)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Fail("Should not have called the rest client getter factory")
				return nil
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Fail("Should not have called the helm uninstaller factory")
				return nil, nil
			},
			configLookup,
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(deregister.FailedToFindClusterCredentials(testErr, remoteClusterName)))
	})

	It("responds with the appropriate error if CSR uninstallation fails", func() {
		testErr := eris.New("test-err")
		helmUninstaller := mock_helm_uninstall.NewMockUninstaller(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)
		configLookup := mock_config_lookup.NewMockKubeConfigLookup(ctrl)
		kubeConfigSecretRef := &zephyr_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.GetWriteNamespace(),
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &zephyr_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.GetWriteNamespace(),
			},
			Spec: zephyr_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		helmUninstaller.EXPECT().
			Run(cliconstants.CsrAgentReleaseName).
			Return(nil, testErr)
		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(&kube.ConvertedConfigs{RestConfig: remoteRestConfig}, nil)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Expect(cfg).To(Equal(remoteRestConfig))
				return restClientGetter
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Expect(namespace).To(Equal(remoteWriteNamespace))

				return helmUninstaller, nil
			},
			configLookup,
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(deregister.FailedToUninstallCsrAgent(testErr, remoteClusterName)))
	})

	It("responds with the appropriate error if CRD removal fails", func() {
		testErr := eris.New("test-err")
		helmUninstaller := mock_helm_uninstall.NewMockUninstaller(ctrl)
		restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		configLookup := mock_config_lookup.NewMockKubeConfigLookup(ctrl)
		kubeConfigSecretRef := &zephyr_core_types.ResourceRef{
			Name:      "kube-config-secret",
			Namespace: env.GetWriteNamespace(),
		}
		remoteClusterName := "remote-cluster-name"
		remoteWriteNamespace := "remote-write-namespace"
		clusterToDeregister := &zephyr_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      remoteClusterName,
				Namespace: env.GetWriteNamespace(),
			},
			Spec: zephyr_discovery_types.KubernetesClusterSpec{
				SecretRef:      kubeConfigSecretRef,
				WriteNamespace: remoteWriteNamespace,
			},
		}
		helmUninstaller.EXPECT().
			Run(cliconstants.CsrAgentReleaseName).
			Return(nil, nil)
		configLookup.EXPECT().
			FromCluster(ctx, clusterToDeregister.GetName()).
			Return(&kube.ConvertedConfigs{RestConfig: remoteRestConfig}, nil)
		crdRemover.EXPECT().
			RemoveZephyrCrds(ctx, remoteClusterName, remoteRestConfig).
			Return(false, testErr)

		clusterDeregistrationClient := deregister.NewClusterDeregistrationClient(
			crdRemover,
			func(cfg *rest.Config) resource.RESTClientGetter {
				Expect(cfg).To(Equal(remoteRestConfig))
				return restClientGetter
			},
			func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (uninstaller helm_uninstall.Uninstaller, err error) {
				Expect(namespace).To(Equal(remoteWriteNamespace))

				return helmUninstaller, nil
			},
			configLookup,
		)

		err := clusterDeregistrationClient.Run(ctx, clusterToDeregister)
		Expect(err).To(testutils.HaveInErrorChain(deregister.FailedToRemoveCrds(testErr, remoteClusterName)))
	})
})
