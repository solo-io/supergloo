package istio_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	mp_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	mp_v1alpha1_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	mock_docker "github.com/solo-io/mesh-projects/pkg/common/docker/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/istio"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Istio Mesh Finder", func() {
	var (
		ctrl    *gomock.Controller
		ctx     = context.TODO()
		istioNs = "istio-system"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not detect Istio when it is not there", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		istioMeshFinder := istio.NewIstioMeshFinder(imageParser)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "test-image",
							},
						},
					},
				},
			},
		}

		mesh, err := istioMeshFinder.ScanDeployment(ctx, deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("reports an error when the image name is unparseable", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		istioMeshFinder := istio.NewIstioMeshFinder(imageParser)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, ClusterName: "test-cluster", Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "istio-pilot:latest",
							},
						},
					},
				},
			},
		}

		testErr := eris.New("test-err")

		imageParser.
			EXPECT().
			Parse("istio-pilot:latest").
			Return(nil, testErr)

		mesh, err := istioMeshFinder.ScanDeployment(ctx, deployment)

		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(mesh).To(BeNil())
	})

	It("discovers istio:latest as latest", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		istioMeshFinder := istio.NewIstioMeshFinder(imageParser)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: istioNs, ClusterName: "test-cluster", Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "istio-pilot:latest",
							},
						},
					},
				},
			},
		}

		expectedMesh := &mp_v1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-istio-system-test-cluster",
				Namespace: env.DefaultWriteNamespace,
				Labels:    istio.DiscoveryLabels,
			},
			Spec: mp_v1alpha1_types.MeshSpec{
				MeshType: &mp_v1alpha1_types.MeshSpec_Istio{
					Istio: &mp_v1alpha1_types.IstioMesh{
						Installation: &mp_v1alpha1_types.MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               "latest",
						},
					},
				},
				Cluster: &mp_v1alpha1_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: env.DefaultWriteNamespace,
				},
			},
		}

		imageParser.
			EXPECT().
			Parse("istio-pilot:latest").
			Return(&docker.Image{
				Domain: "test.com",
				Path:   "istio",
				Tag:    "latest",
			}, nil)

		mesh, err := istioMeshFinder.ScanDeployment(ctx, deployment)

		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})
})
