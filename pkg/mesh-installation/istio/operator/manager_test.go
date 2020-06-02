package operator_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	mock_docker "github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
	mock_operator "github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator/mocks"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Istio Operator Manager", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can generate the manifest and apply it", func() {
		operatorDao := mock_operator.NewMockOperatorDao(ctrl)
		manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
		imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
		operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

		deploymentManifest := "deployment-manifest-yaml"
		operatorConfigManifest := "operator-config-manifest.yaml"

		manifestBuilder.EXPECT().
			BuildOperatorDeploymentManifest(operator.IstioVersion1_5, operator.DefaultIstioOperatorNamespace, false).
			Return(deploymentManifest, nil)
		manifestBuilder.EXPECT().
			BuildOperatorConfigurationWithProfile("my-test-profile", operator.DefaultIstioOperatorNamespace).
			Return(operatorConfigManifest, nil)
		operatorDao.EXPECT().
			ApplyManifest(operator.DefaultIstioOperatorNamespace, deploymentManifest).
			Return(nil)
		operatorDao.EXPECT().
			ApplyManifest(operator.DefaultIstioOperatorNamespace, operatorConfigManifest).
			Return(nil)
		operatorDao.EXPECT().
			FindOperatorDeployment(operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace).
			Return(nil, nil)

		err := operatorManager.InstallOperatorApplication(&operator.InstallationOptions{
			IstioVersion:        operator.IstioVersion1_5,
			InstallationProfile: "my-test-profile",
			InstallNamespace:    operator.DefaultIstioOperatorNamespace,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("fails if generating the manifest fails", func() {
		operatorDao := mock_operator.NewMockOperatorDao(ctrl)
		manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
		imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
		operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

		testErr := eris.New("test-err")

		manifestBuilder.EXPECT().
			BuildOperatorDeploymentManifest(operator.IstioVersion1_5, operator.DefaultIstioOperatorNamespace, false).
			Return("", testErr)

		err := operatorManager.InstallOperatorApplication(&operator.InstallationOptions{
			IstioVersion:        operator.IstioVersion1_5,
			InstallationProfile: "my-test-profile",
			InstallNamespace:    operator.DefaultIstioOperatorNamespace,
		})
		Expect(err).To(testutils.HaveInErrorChain(operator.FailedToGenerateInstallManifest(testErr)))
	})

	It("only writes the operator config if the deployment exists at an acceptable version", func() {
		operatorDao := mock_operator.NewMockOperatorDao(ctrl)
		manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
		imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
		operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

		deploymentManifest := "deployment-manifest-yaml"
		operatorConfigManifest := "operator-config-manifest.yaml"
		imageName := "istio-operator"
		imageVersion := "1.5.1"
		image := imageName + ":" + imageVersion

		manifestBuilder.EXPECT().
			BuildOperatorDeploymentManifest(operator.IstioVersion1_5, operator.DefaultIstioOperatorNamespace, false).
			Return(deploymentManifest, nil)
		manifestBuilder.EXPECT().
			BuildOperatorConfigurationWithProfile("my-test-profile", operator.DefaultIstioOperatorNamespace).
			Return(operatorConfigManifest, nil)
		operatorDao.EXPECT().
			FindOperatorDeployment(operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace).
			Return(&k8s_apps_v1.Deployment{
				Spec: k8s_apps_v1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{{
								Image: image,
							}},
						},
					},
				},
			}, nil)
		imageNameParser.EXPECT().
			Parse(image).
			Return(&docker.Image{
				Tag: imageVersion,
			}, nil)
		operatorDao.EXPECT().
			ApplyManifest(operator.DefaultIstioOperatorNamespace, operatorConfigManifest).
			Return(nil)

		err := operatorManager.InstallOperatorApplication(&operator.InstallationOptions{
			IstioVersion:        operator.IstioVersion1_5,
			InstallationProfile: "my-test-profile",
			InstallNamespace:    operator.DefaultIstioOperatorNamespace,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("fails if the operator is at a version that is too old", func() {
		operatorDao := mock_operator.NewMockOperatorDao(ctrl)
		manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
		imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
		operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

		deploymentManifest := "deployment-manifest-yaml"
		imageName := "istio-operator"
		imageVersion := "1.0.0"
		image := imageName + ":" + imageVersion

		manifestBuilder.EXPECT().
			BuildOperatorDeploymentManifest(operator.IstioVersion1_5, operator.DefaultIstioOperatorNamespace, false).
			Return(deploymentManifest, nil)
		operatorDao.EXPECT().
			FindOperatorDeployment(operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace).
			Return(&k8s_apps_v1.Deployment{
				Spec: k8s_apps_v1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{{
								Image: image,
							}},
						},
					},
				},
			}, nil)
		imageNameParser.EXPECT().
			Parse(image).
			Return(&docker.Image{
				Tag: imageVersion,
			}, nil)

		err := operatorManager.InstallOperatorApplication(&operator.InstallationOptions{
			IstioVersion:        operator.IstioVersion1_5,
			InstallationProfile: "my-test-profile",
			InstallNamespace:    operator.DefaultIstioOperatorNamespace,
		})
		Expect(err).To(HaveOccurred())
	})
})
