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

	When("installing Istio", func() {
		It("can generate the manifest and apply it", func() {
			operatorDao := mock_operator.NewMockOperatorDao(ctrl)
			manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
			imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
			operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

			deploymentManifest := "deployment-manifest-yaml"
			operatorConfigManifest := "operator-config-manifest.yaml"
			totalManifest := `
---
` + deploymentManifest + `
---
` + operatorConfigManifest

			manifestBuilder.EXPECT().
				BuildOperatorDeploymentManifest(operator.IstioVersion1_5, operator.DefaultIstioOperatorNamespace, false).
				Return(deploymentManifest, nil)
			manifestBuilder.EXPECT().
				BuildOperatorConfigurationWithProfile("my-test-profile", operator.DefaultIstioOperatorNamespace).
				Return(operatorConfigManifest, nil)
			operatorDao.EXPECT().
				ApplyManifest(operator.DefaultIstioOperatorNamespace, totalManifest).
				Return(nil)

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
	})

	When("validating any existing operator", func() {
		When("the operator does not exist", func() {
			It("indicates an install is needed", func() {
				operatorDao := mock_operator.NewMockOperatorDao(ctrl)
				manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
				imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
				operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

				operatorDao.EXPECT().
					FindOperatorDeployment(operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace).
					Return(nil, nil)

				installNeeded, err := operatorManager.ValidateOperatorNamespace(operator.IstioVersion1_5, operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace, "test-cluster")
				Expect(err).NotTo(HaveOccurred())
				Expect(installNeeded).To(BeTrue(), "An install should be required")
			})

		})

		When("the operator is already installed", func() {
			Context("and its version is the requested version", func() {
				It("indicates an install is not needed", func() {
					operatorDao := mock_operator.NewMockOperatorDao(ctrl)
					manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
					imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
					operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

					imageName := "istio-operator"
					imageVersion := "1.5.1"
					image := imageName + ":" + imageVersion

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

					installNeeded, err := operatorManager.ValidateOperatorNamespace(operator.IstioVersion1_5, operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace, "test-cluster")
					Expect(err).NotTo(HaveOccurred())
					Expect(installNeeded).To(BeFalse(), "An install should not be required")
				})
			})

			Context("and its version is not the requested version", func() {
				It("fails validation if the operator is outdated", func() {
					operatorDao := mock_operator.NewMockOperatorDao(ctrl)
					manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
					imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
					operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

					imageName := "istio-operator"
					imageVersion := "1.4.0"
					image := imageName + ":" + imageVersion

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

					_, err := operatorManager.ValidateOperatorNamespace(operator.IstioVersion1_5, operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace, "test-cluster")
					Expect(err).To(testutils.HaveInErrorChain(operator.IncompatibleOperatorVersion("v1.4.0", "v1.5.1")))
				})

				It("indicates an install is not needed if the operator is more recent than requested", func() {
					operatorDao := mock_operator.NewMockOperatorDao(ctrl)
					manifestBuilder := mock_operator.NewMockInstallerManifestBuilder(ctrl)
					imageNameParser := mock_docker.NewMockImageNameParser(ctrl)
					operatorManager := operator.NewManager(operatorDao, manifestBuilder, imageNameParser)

					imageName := "istio-operator"
					imageVersion := "1.5.100"
					image := imageName + ":" + imageVersion

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

					installNeeded, err := operatorManager.ValidateOperatorNamespace(operator.IstioVersion1_5, operator.DefaultIstioOperatorDeploymentName, operator.DefaultIstioOperatorNamespace, "test-cluster")
					Expect(err).NotTo(HaveOccurred())
					Expect(installNeeded).To(BeFalse(), "An install should not be required")
				})
			})
		})
	})
})
