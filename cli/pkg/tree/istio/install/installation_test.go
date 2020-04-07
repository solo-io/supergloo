package install_test

import (
	"context"
	"fmt"
	"os"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	mock_kube "github.com/solo-io/service-mesh-hub/cli/pkg/common/kube/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	installcmd "github.com/solo-io/service-mesh-hub/cli/pkg/tree/istio/install"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/istio/operator"
	mock_operator "github.com/solo-io/service-mesh-hub/cli/pkg/tree/istio/operator/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server"
	mock_server "github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	mock_docker "github.com/solo-io/service-mesh-hub/pkg/common/docker/mocks"
	mock_cli_runtime "github.com/solo-io/service-mesh-hub/test/mocks/cli_runtime"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type necessaryMocks struct {
	kubeLoader             *cli_mocks.MockKubeLoader
	restClientGetter       *mock_cli_runtime.MockRESTClientGetter
	unstructuredKubeClient *mock_kube.MockUnstructuredKubeClient
	operatorManager        *mock_operator.MockOperatorManager
	deploymentClient       *mock_server.MockDeploymentClient
	imageNameParser        *mock_docker.MockImageNameParser
	manifestBuilder        *mock_operator.MockInstallerManifestBuilder
	fileReader             *cli_mocks.MockFileReader

	meshctl cli_test.MockMeshctl
}

var _ = Describe("Istio installation", func() {
	var (
		ctrl           *gomock.Controller
		testErr        = eris.New("test-err")
		clusterName    = "cluster-name"
		kubeConfigPath = "/fake/path/to/kubeconfig"
		ctx            = context.TODO()
		setupMocks     = func() necessaryMocks {
			kubeLoader := cli_mocks.NewMockKubeLoader(ctrl)
			restClientGetter := mock_cli_runtime.NewMockRESTClientGetter(ctrl)

			kubeLoader.EXPECT().RESTClientGetter(kubeConfigPath, "").Return(restClientGetter)
			kubeLoader.EXPECT().GetRawConfigForContext(kubeConfigPath, "").Return(clientcmdapi.Config{
				CurrentContext: "minikube",
				Contexts: map[string]*clientcmdapi.Context{
					"minikube": {
						Cluster: clusterName,
					},
				},
			}, nil)

			unstructuredClient := mock_kube.NewMockUnstructuredKubeClient(ctrl)
			operatorManager := mock_operator.NewMockOperatorManager(ctrl)
			deploymentClient := mock_server.NewMockDeploymentClient(ctrl)
			imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
			operatorManifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
			fileReader := cli_mocks.NewMockFileReader(ctrl)

			unstructuredClientFactory := func(_ resource.RESTClientGetter) kube.UnstructuredKubeClient {
				return unstructuredClient
			}

			operatorManagerFactory := func(
				_ kube.UnstructuredKubeClient,
				_ operator.InstallerManifestBuilder,
				_ server.DeploymentClient,
				_ docker.ImageNameParser,
				_ *options.IstioInstallationConfig,
			) operator.OperatorManager {
				return operatorManager
			}

			meshctl := cli_test.MockMeshctl{
				MockController: ctrl,
				Ctx:            ctx,
				Clients: common.Clients{
					UnstructuredKubeClientFactory: unstructuredClientFactory,
					IstioClients: common.IstioClients{
						OperatorManagerFactory:  operatorManagerFactory,
						OperatorManifestBuilder: operatorManifestBuilder,
					},
					DeploymentClient: deploymentClient,
				},
				KubeLoader:      kubeLoader,
				ImageNameParser: imageNameParser,
				FileReader:      fileReader,
			}

			return necessaryMocks{
				kubeLoader:             kubeLoader,
				restClientGetter:       restClientGetter,
				unstructuredKubeClient: unstructuredClient,
				operatorManager:        operatorManager,
				deploymentClient:       deploymentClient,
				imageNameParser:        imageNameParser,
				manifestBuilder:        operatorManifestBuilder,
				fileReader:             fileReader,
				meshctl:                meshctl,
			}
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		os.Setenv("KUBECONFIG", kubeConfigPath)
	})

	AfterEach(func() {
		ctrl.Finish()
		os.Unsetenv("KUBECONFIG")
	})

	It("can install Istio in its demo profile through installing the operator", func() {
		mocks := setupMocks()

		demoControlPlaneSpec := "demo-control-plane"

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)
		mocks.manifestBuilder.EXPECT().GetOperatorSpecWithProfile("demo", cliconstants.DefaultIstioOperatorNamespace).Return(demoControlPlaneSpec, nil)

		demoControlPlaneResource := []*resource.Info{{
			Name: "demo-control-plane",
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "IstioOperator",
				},
			},
		}}

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, demoControlPlaneSpec).
			Return(demoControlPlaneResource, nil)

		mocks.unstructuredKubeClient.
			EXPECT().
			Create(cliconstants.DefaultIstioOperatorNamespace, demoControlPlaneResource).
			Return(nil, nil)

		output, err := mocks.meshctl.Invoke("istio install --profile=demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(`Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'

The IstioOperator has been written to cluster 'cluster-name' in namespace 'istio-operator'. The Istio operator should process it momentarily and install Istio.
`))
	})

	It("outputs the operator install manifest in dry-run mode", func() {
		mocks := setupMocks()

		installationManifest := "now THIS is pod racing"

		mocks.manifestBuilder.EXPECT().
			Build(&options.IstioInstallationConfig{
				CreateNamespace:        true,
				CreateIstioOperatorCRD: true,
				InstallNamespace:       cliconstants.DefaultIstioOperatorNamespace,
			}).
			Return(installationManifest, nil)

		output, err := mocks.meshctl.Invoke("istio install --dry-run")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(installationManifest + "\n"))
	})

	It("outputs the operator install manifest and the pre-configured IstioOperator spec, when provided, in dry-run mode", func() {
		mocks := setupMocks()

		installationManifest := "install-operator"
		controlPlaneSpec := "control-plane"

		mocks.manifestBuilder.EXPECT().
			Build(&options.IstioInstallationConfig{
				CreateNamespace:        true,
				CreateIstioOperatorCRD: true,
				InstallNamespace:       cliconstants.DefaultIstioOperatorNamespace,
			}).
			Return(installationManifest, nil)

		mocks.manifestBuilder.EXPECT().
			GetOperatorSpecWithProfile("demo", cliconstants.DefaultIstioOperatorNamespace).
			Return(controlPlaneSpec, nil)

		output, err := mocks.meshctl.Invoke("istio install --dry-run --profile=demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(installationManifest + "\n\n---\n" + controlPlaneSpec + "\n"))
	})

	It("outputs the operator install manifest and the user-provided IstioOperator in dry-run mode", func() {
		mocks := setupMocks()

		installationManifest := "install-operator"
		controlPlaneSpec := "control-plane"

		mocks.manifestBuilder.EXPECT().
			Build(&options.IstioInstallationConfig{
				CreateNamespace:        true,
				CreateIstioOperatorCRD: true,
				InstallNamespace:       cliconstants.DefaultIstioOperatorNamespace,
			}).
			Return(installationManifest, nil)

		mocks.fileReader.EXPECT().
			Exists("/foo/bar").
			Return(true, nil)

		mocks.fileReader.EXPECT().
			Read("/foo/bar").
			Return([]byte(controlPlaneSpec), nil)

		output, err := mocks.meshctl.Invoke("istio install --dry-run --operator-spec=/foo/bar")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(installationManifest + "\n\n---\n" + controlPlaneSpec + "\n"))
	})

	It("reads an IstioOperator spec provided by the user", func() {
		mocks := setupMocks()

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)

		specFile := "/path/to/spec/file"
		specContent := "this is my IstioOperator spec"

		mocks.fileReader.EXPECT().Exists(specFile).Return(true, nil)
		mocks.fileReader.EXPECT().Read(specFile).Return([]byte(specContent), nil)

		controlPlaneResource := []*resource.Info{{
			Name: "custom-control-plane",
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "IstioOperator",
				},
			},
		}}

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, specContent).
			Return(controlPlaneResource, nil)

		mocks.unstructuredKubeClient.
			EXPECT().
			Create(cliconstants.DefaultIstioOperatorNamespace, controlPlaneResource).
			Return(nil, nil)

		output, err := mocks.meshctl.Invoke(fmt.Sprintf("istio install --operator-spec %s", specFile))

		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(`Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'

The IstioOperator has been written to cluster 'cluster-name' in namespace 'istio-operator'. The Istio operator should process it momentarily and install Istio.
`))
	})

	It("reads an IstioControlPlane spec provided by the via stdin", func() {
		mocks := setupMocks()

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)

		specContent := "this is my IstioControlPlane spec"
		mocks.meshctl.Stdin = specContent

		controlPlaneResource := []*resource.Info{{
			Name: "custom-control-plane",
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "IstioOperator",
				},
			},
		}}

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, specContent).
			Return(controlPlaneResource, nil)

		mocks.unstructuredKubeClient.
			EXPECT().
			Create(cliconstants.DefaultIstioOperatorNamespace, controlPlaneResource).
			Return(nil, nil)

		output, err := mocks.meshctl.Invoke("istio install --operator-spec=-")

		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(`Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'

The IstioOperator has been written to cluster 'cluster-name' in namespace 'istio-operator'. The Istio operator should process it momentarily and install Istio.
`))
	})

	It("can install the operator without providing an IstioOperator spec", func() {
		mocks := setupMocks()

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)

		output, err := mocks.meshctl.Invoke(fmt.Sprintf("istio install"))

		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(`Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'

The Istio operator has been installed to cluster 'cluster-name' in namespace 'istio-operator'. No IstioOperator custom resource was provided to meshctl, so Istio is currently not fully installed yet. Write a IstioOperator CR to cluster 'cluster-name' to complete your installation
`))
	})

	It("reports an error if you specify both --profile=demo and --operator-spec", func() {
		mocks := setupMocks()

		specFile := "/path/to/spec/file"

		output, err := mocks.meshctl.Invoke(fmt.Sprintf("istio install --profile=demo --operator-spec %s", specFile))

		Expect(err).To(testutils.HaveInErrorChain(installcmd.ConflictingControlPlaneSettings))
		Expect(output).To(BeEmpty())
	})

	It("can use an already-existing operator to install Istio in its demo profile", func() {
		mocks := setupMocks()

		demoControlPlaneSpec := "demo-control-plane"

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(false, nil)
		mocks.manifestBuilder.EXPECT().GetOperatorSpecWithProfile("demo", cliconstants.DefaultIstioOperatorNamespace).Return(demoControlPlaneSpec, nil)

		demoControlPlaneResource := []*resource.Info{{
			Name: "demo-control-plane",
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "IstioOperator",
				},
			},
		}}

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, demoControlPlaneSpec).
			Return(demoControlPlaneResource, nil)

		mocks.unstructuredKubeClient.
			EXPECT().
			Create(cliconstants.DefaultIstioOperatorNamespace, demoControlPlaneResource).
			Return(nil, nil)

		output, err := mocks.meshctl.Invoke("istio install --profile=demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(`The Istio operator is already installed to cluster 'cluster-name' in namespace 'istio-operator' and is suitable for use. Continuing with the Istio installation.

The IstioOperator has been written to cluster 'cluster-name' in namespace 'istio-operator'. The Istio operator should process it momentarily and install Istio.
`))
	})

	It("reports the appropriate error if the provided IstioOperator is unparseable", func() {
		mocks := setupMocks()

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)

		specFile := "/path/to/spec/file"
		specContent := "there's always a bigger fish"

		mocks.fileReader.EXPECT().Exists(specFile).Return(true, nil)
		mocks.fileReader.EXPECT().Read(specFile).Return([]byte(specContent), nil)

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, specContent).
			Return(nil, testErr)

		output, err := mocks.meshctl.Invoke(fmt.Sprintf("istio install --operator-spec %s", specFile))

		Expect(err).To(testutils.HaveInErrorChain(installcmd.FailedToParseControlPlaneSettings(testErr)))
		Expect(output).To(Equal("Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'\n"))
	})

	It("reports an error if there are multiple resources in the user-provided IstioOperator manifest", func() {
		mocks := setupMocks()

		demoControlPlaneSpec := "demo-control-plane"

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)
		mocks.manifestBuilder.EXPECT().GetOperatorSpecWithProfile("demo", cliconstants.DefaultIstioOperatorNamespace).Return(demoControlPlaneSpec, nil)

		demoControlPlaneResource := []*resource.Info{
			{
				Name: "demo-control-plane",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "IstioOperator",
					},
				},
			},
			{
				Name: "some-other-resource",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "kind-doesnt-matter-here",
					},
				},
			},
		}

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, demoControlPlaneSpec).
			Return(demoControlPlaneResource, nil)

		output, err := mocks.meshctl.Invoke("istio install --profile=demo")
		Expect(err).To(testutils.HaveInErrorChain(installcmd.TooManyControlPlaneResources(len(demoControlPlaneResource))))
		Expect(output).To(Equal("Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'\n"))
	})

	It("reports an error if the user provides some resource other than an IstioOperator", func() {
		mocks := setupMocks()

		demoControlPlaneSpec := "demo-control-plane"

		mocks.operatorManager.EXPECT().ValidateOperatorNamespace(clusterName).Return(true, nil)
		mocks.operatorManager.EXPECT().Install().Return(nil)
		mocks.manifestBuilder.EXPECT().GetOperatorSpecWithProfile("demo", cliconstants.DefaultIstioOperatorNamespace).Return(demoControlPlaneSpec, nil)

		demoControlPlaneResource := []*resource.Info{
			{
				Name: "some-other-resource",
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "whoops-wrong-kind",
					},
				},
			},
		}

		mocks.unstructuredKubeClient.
			EXPECT().
			BuildResources(cliconstants.DefaultIstioOperatorNamespace, demoControlPlaneSpec).
			Return(demoControlPlaneResource, nil)

		output, err := mocks.meshctl.Invoke("istio install --profile=demo")
		Expect(err).To(testutils.HaveInErrorChain(installcmd.UnknownControlPlaneKind("whoops-wrong-kind")))
		Expect(output).To(Equal("Installing the Istio operator to cluster 'cluster-name' in namespace 'istio-operator'\n"))
	})
})
