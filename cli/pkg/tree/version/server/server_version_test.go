package server_test

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	mock_server "github.com/solo-io/mesh-projects/cli/pkg/tree/version/server/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ServerVersion", func() {
	var (
		ctrl                *gomock.Controller
		kubeConfigClient    *mock_server.MockDeploymentClient
		serverVersionClient server.ServerVersionClient
		namespace           = "test-namespace"
		labelSelector       = "app=sm-marketplace"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		kubeConfigClient = mock_server.NewMockDeploymentClient(ctrl)
		serverVersionClient = server.NewServerVersionClient(namespace, kubeConfigClient)
	})

	It("GetServerVersion should parse image repository, registry, and tag correctly", func() {
		image1 := "gcr.io/service-mesh-hub/foo/bar/mesh-discovery:latest"
		image2 := "gcr.io/service-mesh-hub/boo/baz/mesh-discovery:latest"
		image3 := "gcr.io/service-mesh-hub/gloo/foo/mesh-discovery:latest"
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
			Namespace: namespace,
			Containers: []*server.ImageMeta{
				{
					Tag:      "latest",
					Name:     "gcr.io/service-mesh-hub/foo/bar/mesh-discovery",
					Registry: "gcr.io/service-mesh-hub/foo/bar",
				},
				{
					Tag:      "latest",
					Name:     "gcr.io/service-mesh-hub/boo/baz/mesh-discovery",
					Registry: "gcr.io/service-mesh-hub/boo/baz",
				},
				{
					Tag:      "latest",
					Name:     "gcr.io/service-mesh-hub/gloo/foo/mesh-discovery",
					Registry: "gcr.io/service-mesh-hub/gloo/foo",
				},
			},
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(serverVersion.Namespace).To(Equal(expectedServerVersion.Namespace))
		Expect(serverVersion.Containers).To(ConsistOf(expectedServerVersion.Containers))
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
