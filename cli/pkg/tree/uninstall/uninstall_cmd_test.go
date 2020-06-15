package uninstall_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	mock_kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	mock_types "github.com/solo-io/go-utils/installutils/helminstall/types/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	mock_cluster_registration "github.com/solo-io/service-mesh-hub/pkg/common/cluster-registration/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mock_crd_uninstall "github.com/solo-io/service-mesh-hub/pkg/common/kube/crd/mocks"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig/mocks"
	k8s_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Crd Uninstaller", func() {
	var (
		ctrl          *gomock.Controller
		ctx           context.Context
		masterRestCfg = &rest.Config{
			Host: "arbitrary.com",
		}
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can uninstall everything except the namespace by default", func() {
		kubeLoader := mock_kubeconfig.NewMockKubeLoader(ctrl)
		helmClient := mock_types.NewMockHelmClient(ctrl)
		helmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		kubeClusterClient := mock_smh_discovery.NewMockKubernetesClusterClient(ctrl)
		clusterDeregistrationClient := mock_cluster_registration.NewMockClusterDeregistrationClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		releaseName := constants.ServiceMeshHubReleaseName
		cluster1 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-1",
			},
		}
		cluster2 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-2",
			},
		}

		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(masterRestCfg, nil)
		helmClient.EXPECT().
			NewUninstall(container_runtime.GetWriteNamespace()).
			Return(helmUninstaller, nil)
		helmUninstaller.EXPECT().
			Run(releaseName).
			Return(nil, nil)
		kubeClusterClient.EXPECT().
			ListKubernetesCluster(ctx, client.InNamespace(container_runtime.GetWriteNamespace())).
			Return(&smh_discovery.KubernetesClusterList{
				Items: []smh_discovery.KubernetesCluster{*cluster1, *cluster2},
			}, nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster1).
			Return(nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster2).
			Return(nil)
		crdRemover.EXPECT().
			RemovesmhCrds(ctx, "management plane cluster", masterRestCfg).
			Return(true, nil)

		meshctl := cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			KubeClients: common.KubeClients{
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return helmClient
				},
				KubeClusterClient:           kubeClusterClient,
				ClusterDeregistrationClient: clusterDeregistrationClient,
				NamespaceClient:             namespaceClient,
				UninstallClients: common.UninstallClients{
					CrdRemover: crdRemover,
				},
			},
			KubeLoader: kubeLoader,
		}

		stdout, err := meshctl.Invoke("uninstall")
		Expect(err).NotTo(HaveOccurred())
		expectedText := `Service Mesh Hub management plane components have been removed...
Starting to de-register 2 cluster(s). This may take a moment...
All clusters have been de-registered...
Service Mesh Hub CRDs have been de-registered from the management plane...

Service Mesh Hub has been uninstalled
`
		Expect(stdout).To(Equal(expectedText))
	})

	It("remove the namespace when so configured", func() {
		kubeLoader := mock_kubeconfig.NewMockKubeLoader(ctrl)
		helmClient := mock_types.NewMockHelmClient(ctrl)
		helmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		kubeClusterClient := mock_smh_discovery.NewMockKubernetesClusterClient(ctrl)
		clusterDeregistrationClient := mock_cluster_registration.NewMockClusterDeregistrationClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		releaseName := constants.ServiceMeshHubReleaseName
		cluster1 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-1",
			},
		}
		cluster2 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-2",
			},
		}

		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(masterRestCfg, nil)
		helmClient.EXPECT().
			NewUninstall(container_runtime.GetWriteNamespace()).
			Return(helmUninstaller, nil)
		helmUninstaller.EXPECT().
			Run(releaseName).
			Return(nil, nil)
		kubeClusterClient.EXPECT().
			ListKubernetesCluster(ctx, client.InNamespace(container_runtime.GetWriteNamespace())).
			Return(&smh_discovery.KubernetesClusterList{
				Items: []smh_discovery.KubernetesCluster{*cluster1, *cluster2},
			}, nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster1).
			Return(nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster2).
			Return(nil)
		ns := &k8s_core.Namespace{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: container_runtime.GetWriteNamespace(),
			},
		}
		namespaceClient.EXPECT().
			GetNamespace(ctx, container_runtime.GetWriteNamespace()).
			Return(ns, nil)
		namespaceClient.EXPECT().
			DeleteNamespace(ctx, ns.GetName()).
			Return(nil)
		crdRemover.EXPECT().
			RemovesmhCrds(ctx, "management plane cluster", masterRestCfg).
			Return(true, nil)

		meshctl := cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			KubeClients: common.KubeClients{
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return helmClient
				},
				KubeClusterClient:           kubeClusterClient,
				ClusterDeregistrationClient: clusterDeregistrationClient,
				NamespaceClient:             namespaceClient,
				UninstallClients: common.UninstallClients{
					CrdRemover: crdRemover,
				},
			},
			KubeLoader: kubeLoader,
		}

		stdout, err := meshctl.Invoke("uninstall --remove-namespace")
		Expect(err).NotTo(HaveOccurred())
		expectedText := `Service Mesh Hub management plane components have been removed...
Starting to de-register 2 cluster(s). This may take a moment...
All clusters have been de-registered...
Service Mesh Hub management plane namespace has been removed...
Service Mesh Hub CRDs have been de-registered from the management plane...

Service Mesh Hub has been uninstalled
`
		Expect(stdout).To(Equal(expectedText))
	})

	It("is a no-op with a 0 exit code if everything has been uninstalled already", func() {
		kubeLoader := mock_kubeconfig.NewMockKubeLoader(ctrl)
		helmClient := mock_types.NewMockHelmClient(ctrl)
		helmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		kubeClusterClient := mock_smh_discovery.NewMockKubernetesClusterClient(ctrl)
		clusterDeregistrationClient := mock_cluster_registration.NewMockClusterDeregistrationClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		releaseName := constants.ServiceMeshHubReleaseName

		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(masterRestCfg, nil)
		helmClient.EXPECT().
			NewUninstall(container_runtime.GetWriteNamespace()).
			Return(helmUninstaller, nil)
		helmUninstaller.EXPECT().
			Run(releaseName).
			Return(nil, eris.New(uninstall.ReleaseNotFoundHelmErrorMessage))
		kubeClusterClient.EXPECT().
			ListKubernetesCluster(ctx, client.InNamespace(container_runtime.GetWriteNamespace())).
			Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
		namespaceClient.EXPECT().
			GetNamespace(ctx, container_runtime.GetWriteNamespace()).
			Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
		crdRemover.EXPECT().
			RemovesmhCrds(ctx, "management plane cluster", masterRestCfg).
			Return(false, nil)

		meshctl := cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			KubeClients: common.KubeClients{
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return helmClient
				},
				KubeClusterClient:           kubeClusterClient,
				ClusterDeregistrationClient: clusterDeregistrationClient,
				NamespaceClient:             namespaceClient,
				UninstallClients: common.UninstallClients{
					CrdRemover: crdRemover,
				},
			},
			KubeLoader: kubeLoader,
		}

		stdout, err := meshctl.Invoke("uninstall --remove-namespace")
		Expect(err).NotTo(HaveOccurred())
		expectedText := `Management plane components are not running here...
No clusters to deregister...
No CRDs to remove from the management plane cluster...

Service Mesh Hub has been uninstalled
`
		Expect(stdout).To(Equal(expectedText))
	})

	It("can continue on through all the stages even when intermediate ones fail", func() {
		errorNumber := 0
		generateNewErr := func() error {
			errorNumber++
			return eris.Errorf("test-err-%d", errorNumber)
		}
		kubeLoader := mock_kubeconfig.NewMockKubeLoader(ctrl)
		helmClient := mock_types.NewMockHelmClient(ctrl)
		helmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		kubeClusterClient := mock_smh_discovery.NewMockKubernetesClusterClient(ctrl)
		clusterDeregistrationClient := mock_cluster_registration.NewMockClusterDeregistrationClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		releaseName := constants.ServiceMeshHubReleaseName

		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(masterRestCfg, nil)
		helmClient.EXPECT().
			NewUninstall(container_runtime.GetWriteNamespace()).
			Return(helmUninstaller, nil)
		helmUninstaller.EXPECT().
			Run(releaseName).
			Return(nil, generateNewErr())
		kubeClusterClient.EXPECT().
			ListKubernetesCluster(ctx, client.InNamespace(container_runtime.GetWriteNamespace())).
			Return(nil, generateNewErr())
		namespaceClient.EXPECT().
			GetNamespace(ctx, container_runtime.GetWriteNamespace()).
			Return(nil, generateNewErr())
		crdRemover.EXPECT().
			RemovesmhCrds(ctx, "management plane cluster", masterRestCfg).
			Return(true, generateNewErr())

		meshctl := cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			KubeClients: common.KubeClients{
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return helmClient
				},
				KubeClusterClient:           kubeClusterClient,
				ClusterDeregistrationClient: clusterDeregistrationClient,
				NamespaceClient:             namespaceClient,
				UninstallClients: common.UninstallClients{
					CrdRemover: crdRemover,
				},
			},
			KubeLoader: kubeLoader,
		}

		stdout, err := meshctl.Invoke("uninstall --remove-namespace")
		expectedOutput := `Management plane components not removed - Continuing...
` + "\t" + `(test-err-1)
Failed to find registered clusters - Continuing...
` + "\t" + `(test-err-2)
Failed to remove management plane namespace - Continuing...
` + "\t" + `(test-err-3: Failed to remove namespace service-mesh-hub)
Failed to remove CRDs from management plane - Continuing...
` + "\t" + `(test-err-4)

Service Mesh Hub has been uninstalled with errors
`
		Expect(stdout).To(Equal(expectedOutput))
		Expect(err).To(HaveOccurred()) // error doesn't particularly matter here- it's just to get a nonzero exit code
	})

	It("continues when there are clusters to delete but they fail", func() {
		errorNumber := 0
		generateNewErr := func() error {
			errorNumber++
			return eris.Errorf("test-err-%d", errorNumber)
		}
		kubeLoader := mock_kubeconfig.NewMockKubeLoader(ctrl)
		helmClient := mock_types.NewMockHelmClient(ctrl)
		helmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		kubeClusterClient := mock_smh_discovery.NewMockKubernetesClusterClient(ctrl)
		clusterDeregistrationClient := mock_cluster_registration.NewMockClusterDeregistrationClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		releaseName := constants.ServiceMeshHubReleaseName
		cluster1 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-1",
			},
		}
		cluster2 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-2",
			},
		}

		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(masterRestCfg, nil)
		helmClient.EXPECT().
			NewUninstall(container_runtime.GetWriteNamespace()).
			Return(helmUninstaller, nil)
		helmUninstaller.EXPECT().
			Run(releaseName).
			Return(nil, generateNewErr())
		kubeClusterClient.EXPECT().
			ListKubernetesCluster(ctx, client.InNamespace(container_runtime.GetWriteNamespace())).
			Return(&smh_discovery.KubernetesClusterList{
				Items: []smh_discovery.KubernetesCluster{*cluster1, *cluster2},
			}, nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster1).
			Return(generateNewErr())
		namespaceClient.EXPECT().
			GetNamespace(ctx, container_runtime.GetWriteNamespace()).
			Return(nil, generateNewErr())
		crdRemover.EXPECT().
			RemovesmhCrds(ctx, "management plane cluster", masterRestCfg).
			Return(true, generateNewErr())

		meshctl := cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			KubeClients: common.KubeClients{
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return helmClient
				},
				KubeClusterClient:           kubeClusterClient,
				ClusterDeregistrationClient: clusterDeregistrationClient,
				NamespaceClient:             namespaceClient,
				UninstallClients: common.UninstallClients{
					CrdRemover: crdRemover,
				},
			},
			KubeLoader: kubeLoader,
		}

		stdout, err := meshctl.Invoke("uninstall --remove-namespace")
		expectedOutput := `Management plane components not removed - Continuing...
` + "\t" + `(test-err-1)
Starting to de-register 2 cluster(s). This may take a moment...
Failed to de-register all clusters - Continuing...
` + "\t" + `(test-err-2: Failed to de-register cluster cluster-1)
Failed to remove management plane namespace - Continuing...
` + "\t" + `(test-err-3: Failed to remove namespace service-mesh-hub)
Failed to remove CRDs from management plane - Continuing...
` + "\t" + `(test-err-4)

Service Mesh Hub has been uninstalled with errors
`
		Expect(stdout).To(Equal(expectedOutput))
		Expect(err).To(HaveOccurred()) // error doesn't particularly matter here- it's just to get a nonzero exit code
	})

	It("works when things are installed to a nonstandard namespace and have a different release name", func() {
		kubeLoader := mock_kubeconfig.NewMockKubeLoader(ctrl)
		helmClient := mock_types.NewMockHelmClient(ctrl)
		helmUninstaller := mock_types.NewMockHelmUninstaller(ctrl)
		kubeClusterClient := mock_smh_discovery.NewMockKubernetesClusterClient(ctrl)
		clusterDeregistrationClient := mock_cluster_registration.NewMockClusterDeregistrationClient(ctrl)
		namespaceClient := mock_kubernetes_core.NewMockNamespaceClient(ctrl)
		crdRemover := mock_crd_uninstall.NewMockCrdRemover(ctrl)
		releaseName := "different-release-name"
		smhInstallNamespace := "smh-management-plane-namespace"
		cluster1 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-1",
			},
		}
		cluster2 := &smh_discovery.KubernetesCluster{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: "cluster-2",
			},
		}

		kubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(masterRestCfg, nil)
		helmClient.EXPECT().
			NewUninstall(smhInstallNamespace).
			Return(helmUninstaller, nil)
		helmUninstaller.EXPECT().
			Run(releaseName).
			Return(nil, nil)
		kubeClusterClient.EXPECT().
			ListKubernetesCluster(ctx, client.InNamespace(smhInstallNamespace)).
			Return(&smh_discovery.KubernetesClusterList{
				Items: []smh_discovery.KubernetesCluster{*cluster1, *cluster2},
			}, nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster1).
			Return(nil)
		clusterDeregistrationClient.EXPECT().
			Deregister(ctx, cluster2).
			Return(nil)
		ns := &k8s_core.Namespace{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name: smhInstallNamespace,
			},
		}
		namespaceClient.EXPECT().
			GetNamespace(ctx, smhInstallNamespace).
			Return(ns, nil)
		namespaceClient.EXPECT().
			DeleteNamespace(ctx, ns.GetName()).
			Return(nil)
		crdRemover.EXPECT().
			RemovesmhCrds(ctx, "management plane cluster", masterRestCfg).
			Return(true, nil)

		meshctl := cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			KubeClients: common.KubeClients{
				HelmClientFileConfigFactory: func(kubeConfig, kubeContext string) types.HelmClient {
					return helmClient
				},
				KubeClusterClient:           kubeClusterClient,
				ClusterDeregistrationClient: clusterDeregistrationClient,
				NamespaceClient:             namespaceClient,
				UninstallClients: common.UninstallClients{
					CrdRemover: crdRemover,
				},
			},
			KubeLoader: kubeLoader,
		}

		stdout, err := meshctl.Invoke(fmt.Sprintf("uninstall --remove-namespace -n %s --release-name %s", smhInstallNamespace, releaseName))
		Expect(err).NotTo(HaveOccurred())
		expectedText := `Service Mesh Hub management plane components have been removed...
Starting to de-register 2 cluster(s). This may take a moment...
All clusters have been de-registered...
Service Mesh Hub management plane namespace has been removed...
Service Mesh Hub CRDs have been de-registered from the management plane...

Service Mesh Hub has been uninstalled
`
		Expect(stdout).To(Equal(expectedText))
	})
})
