package server_test

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	mock_server "github.com/solo-io/mesh-projects/cli/pkg/tree/version/server/mocks"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	mock_docker "github.com/solo-io/mesh-projects/pkg/common/docker/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ServerVersion", func() {
	var (
		ctrl                *gomock.Controller
		kubeConfigClient    *mock_server.MockDeploymentClient
		serverVersionClient server.ServerVersionClient
		namespace           = "test-namespace"
		labelSelector       = "app=" + env.DefaultWriteNamespace
		imageNameParser     *mock_docker.MockImageNameParser
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		kubeConfigClient = mock_server.NewMockDeploymentClient(ctrl)
		imageNameParser = mock_docker.NewMockImageNameParser(ctrl)
		serverVersionClient = server.NewServerVersionClient(namespace, kubeConfigClient, imageNameParser)
	})

	It("GetServerVersion should parse image repository, registry, and tag correctly", func() {
		image1 := "gcr.io/service-mesh-hub/foo/bar/mesh-discovery:latest"
		image2 := "gcr.io/service-mesh-hub/boo/baz/mesh-discovery:latest"
		image3 := "gcr.io/service-mesh-hub/gloo/foo/mesh-discovery:latest"

		parsedImages := []*docker.Image{
			{
				Domain: "gcr.io",
				Path:   "service-mesh-hub/foo/bar/mesh-discovery",
				Tag:    "latest",
			},
			{
				Domain: "gcr.io",
				Path:   "service-mesh-hub/boo/baz/mesh-discovery",
				Tag:    "latest",
			},
			{
				Domain: "gcr.io",
				Path:   "service-mesh-hub/gloo/foo/mesh-discovery",
				Tag:    "latest",
			},
		}

		imageNameParser.EXPECT().Parse(image1).Return(parsedImages[0], nil)
		imageNameParser.EXPECT().Parse(image2).Return(parsedImages[1], nil)
		imageNameParser.EXPECT().Parse(image3).Return(parsedImages[2], nil)
		deployments := &appsv1.DeploymentList{
			Items: []appsv1.Deployment{
				{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image1,
									},
								},
							},
						},
					},
				},
				{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image2,
									},
									{
										Image: image3,
									},
								},
							},
						},
					},
				},
			},
		}
		kubeConfigClient.
			EXPECT().
			GetDeployments(namespace, labelSelector).
			Return(deployments, nil)

		serverVersion, err := serverVersionClient.GetServerVersion()
		expectedServerVersion := &server.ServerVersion{
			Namespace:  namespace,
			Containers: parsedImages,
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(serverVersion.Namespace).To(Equal(expectedServerVersion.Namespace))
		Expect(serverVersion.Containers).To(Equal(expectedServerVersion.Containers))
	})

	It("GetServerVersion should throw error if KubeConfigClient returns error", func() {
		kubeConfigClientError := fmt.Errorf("some error")
		kubeConfigClient.
			EXPECT().
			GetDeployments(namespace, labelSelector).
			Return(nil, kubeConfigClientError)
		_, err := serverVersionClient.GetServerVersion()
		Expect(err).To(testutils.HaveInErrorChain(server.ConfigClientError(kubeConfigClientError)))
	})
})
