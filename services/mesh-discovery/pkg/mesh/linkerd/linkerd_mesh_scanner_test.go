package linkerd_test

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
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/linkerd"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Istio Mesh Scanner", func() {
	var (
		ctrl      *gomock.Controller
		ctx       context.Context
		linkerdNs = "linkerd"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not discover linkerd when it is not there", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		scanner := linkerd.NewLinkerdMeshScanner(imageParser)

		deployment := &appsv1.Deployment{
			ObjectMeta: k8s_meta_v1.ObjectMeta{Namespace: linkerdNs, Name: "name doesn't matter in this context"},
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

		mesh, err := scanner.ScanDeployment(ctx, deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("discovers linkerd", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		scanner := linkerd.NewLinkerdMeshScanner(imageParser)

		deployment := &appsv1.Deployment{
			ObjectMeta: k8s_meta_v1.ObjectMeta{Namespace: linkerdNs, ClusterName: "test-cluster", Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "linkerd-io/controller:0.6.9",
							},
						},
					},
				},
			},
		}

		imageParser.
			EXPECT().
			Parse("linkerd-io/controller:0.6.9").
			Return(&docker.Image{
				Domain: "docker.io",
				Path:   "linkerd-io/controller",
				Tag:    "0.6.9",
			}, nil)

		expectedMesh := &mp_v1alpha1.Mesh{
			ObjectMeta: k8s_meta_v1.ObjectMeta{
				Name:      "linkerd-linkerd-test-cluster",
				Namespace: env.DefaultWriteNamespace,
				Labels:    linkerd.DiscoveryLabels,
			},
			Spec: mp_v1alpha1_types.MeshSpec{
				MeshType: &mp_v1alpha1_types.MeshSpec_Linkerd{
					Linkerd: &mp_v1alpha1_types.LinkerdMesh{
						Installation: &mp_v1alpha1_types.MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               "0.6.9",
						},
					},
				},
				Cluster: &mp_v1alpha1_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: env.DefaultWriteNamespace,
				},
			},
		}

		mesh, err := scanner.ScanDeployment(ctx, deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("reports an error when the image name is un-parseable", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		scanner := linkerd.NewLinkerdMeshScanner(imageParser)

		deployment := &appsv1.Deployment{
			ObjectMeta: k8s_meta_v1.ObjectMeta{Namespace: linkerdNs, ClusterName: "test-cluster", Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "linkerd-io/controller:0.6.9",
							},
						},
					},
				},
			},
		}

		testErr := eris.New("test-err")

		imageParser.
			EXPECT().
			Parse("linkerd-io/controller:0.6.9").
			Return(nil, testErr)

		mesh, err := scanner.ScanDeployment(ctx, deployment)
		Expect(mesh).To(BeNil())
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
