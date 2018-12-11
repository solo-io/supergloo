package install_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/supergloo/mock/pkg/kube"
	"github.com/solo-io/supergloo/mock/pkg/secret"
	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install"
	"github.com/solo-io/supergloo/pkg/install/consul"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/pkg/kube"
	"github.com/solo-io/supergloo/test/util"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"context"
	"testing"
)

var T *testing.T

func TestInstallSyncer(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Shared Suite")
}

var _ = Describe("Install syncer", func() {

	const (
		testChartPath        = "testChartPath"
		testMeshName         = "testMeshName"
		testInstallNamespace = "testInstallNamespace"
	)

	var (
		syncer              *install.InstallSyncer
		ctx                 context.Context
		mockCrdClient       *mock_kube.MockCrdClient
		mockNamespaceClient *mock_kube.MockNamespaceClient
		mockRbacClient      *mock_kube.MockRbacClient
		mockSecretSyncer    *mock_secret.MockSecretSyncer

		testError error
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(T)
		defer ctrl.Finish()

		ctx = context.TODO()
		testError = errors.Errorf("test error")

		inMemoryCache := memory.NewInMemoryResourceCache()
		meshClient, err := v1.NewMeshClient(&factory.MemoryResourceClientFactory{
			Cache: inMemoryCache,
		})
		Expect(err).To(BeNil())
		Expect(meshClient.Register()).To(BeNil())

		istioSecretClient, err := istiov1.NewIstioCacertsSecretClient(&factory.MemoryResourceClientFactory{
			Cache: inMemoryCache,
		})
		Expect(err).To(BeNil())
		Expect(istioSecretClient.Register()).To(BeNil())

		mockCrdClient = mock_kube.NewMockCrdClient(ctrl)
		mockNamespaceClient = mock_kube.NewMockNamespaceClient(ctrl)
		mockRbacClient = mock_kube.NewMockRbacClient(ctrl)
		mockSecretSyncer = mock_secret.NewMockSecretSyncer(ctrl)

		syncer, err = install.NewInstallSyncer(meshClient, istioSecretClient, mockSecretSyncer, mockRbacClient, mockNamespaceClient, mockCrdClient)
	})

	getCrds := func() []*v1beta1.CustomResourceDefinition {
		crds, err := kube.CrdsFromManifest(istio.IstioCrdYaml)
		Expect(err).To(BeNil())
		return crds
	}

	getIstioInstall := func() *v1.Install {
		installCrd := util.GetInstallWithoutMeshType(testChartPath, testMeshName, true)
		installCrd.MeshType = &v1.Install_Istio{
			Istio: &v1.Istio{
				InstallationNamespace: testInstallNamespace,
			},
		}
		return installCrd
	}

	getLinkerd2Install := func() *v1.Install {
		installCrd := util.GetInstallWithoutMeshType(testChartPath, testMeshName, true)
		installCrd.MeshType = &v1.Install_Linkerd2{
			Linkerd2: &v1.Linkerd2{
				InstallationNamespace: testInstallNamespace,
			},
		}
		return installCrd
	}

	getConsulInstall := func() *v1.Install {
		installCrd := util.GetInstallWithoutMeshType(testChartPath, testMeshName, true)
		installCrd.MeshType = &v1.Install_Consul{
			Consul: &v1.Consul{
				InstallationNamespace: testInstallNamespace,
			},
		}
		return installCrd
	}

	It("install istio propagates create namespace error", func() {
		snap := util.GetSnapshot(getIstioInstall())
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("Error setting up namespace"))
	})

	It("install consul propagates create namespace error", func() {
		snap := util.GetSnapshot(getConsulInstall())
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("Error setting up namespace"))
	})

	It("install linkerd2 propagates create namespace error", func() {
		snap := util.GetSnapshot(getLinkerd2Install())
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("Error setting up namespace"))
	})

	It("install istio propagates create CRB error", func() {
		snap := util.GetSnapshot(getIstioInstall())
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(nil)
		mockRbacClient.EXPECT().CreateCrbIfNotExist(istio.CrbName, testInstallNamespace).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("Error creating CRB"))
	})

	It("install consul propagates create CRB error", func() {
		snap := util.GetSnapshot(getConsulInstall())
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(nil)
		mockRbacClient.EXPECT().CreateCrbIfNotExist(consul.CrbName, testInstallNamespace).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("Error creating CRB"))
	})

	It("install istio propagates crd creation error", func() {
		snap := util.GetSnapshot(getIstioInstall())
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(nil)
		mockRbacClient.EXPECT().CreateCrbIfNotExist(istio.CrbName, testInstallNamespace).Times(1).Return(nil)
		mockCrdClient.EXPECT().CreateCrds(getCrds()).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("creating istio crds"))
		Expect(err.Error()).To(ContainSubstring("Error doing pre-helm install steps"))

	})

	It("install istio propagates sync secret error", func() {
		install := getIstioInstall()
		snap := util.GetSnapshot(install)
		mockNamespaceClient.EXPECT().CreateNamespaceIfNotExist(testInstallNamespace).Times(1).Return(nil)
		mockRbacClient.EXPECT().CreateCrbIfNotExist(istio.CrbName, testInstallNamespace).Times(1).Return(nil)
		mockCrdClient.EXPECT().CreateCrds(getCrds()).Times(1).Return(nil)
		updatedCtx := contextutils.WithLogger(ctx, "install-syncer")
		mockSecretSyncer.EXPECT().SyncSecret(updatedCtx, testInstallNamespace, install.Encryption, util.GetTestSecrets(), true).Times(1).Return(testError)
		err := syncer.Sync(ctx, snap)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("syncing secret"))
		Expect(err.Error()).To(ContainSubstring("Error doing pre-helm install steps"))
	})

})
