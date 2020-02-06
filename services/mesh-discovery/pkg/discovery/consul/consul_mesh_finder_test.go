package consul_test

import (
	"context"
	"fmt"

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
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/consul"
	mock_consul "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/consul/mocks"
	appsv1 "k8s.io/api/apps/v1"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	consulNs       = "consul-ns"
	consulVersion  = "1.6.2"
	datacenterName = "minidc"
)

var _ = Describe("Consul Mesh Finder", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("doesn't discover consul if it is not present", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)
		consulInstallationFinder := mock_consul.NewMockConsulConnectInstallationFinder(ctrl)

		consulMeshFinder := consul.NewConsulMeshFinder(imageParser, consulInstallationFinder)

		nonConsulDeployment := &k8s_apps_v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{{
							Image: "test-image",
						}},
					},
				},
			},
		}

		consulInstallationFinder.
			EXPECT().
			IsConsulConnect(nonConsulDeployment.Spec.Template.Spec.Containers[0]).
			Return(false, nil)

		mesh, err := consulMeshFinder.ScanDeployment(ctx, nonConsulDeployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("can discover consul", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)
		consulInstallationFinder := mock_consul.NewMockConsulConnectInstallationFinder(ctrl)

		consulMeshFinder := consul.NewConsulMeshFinder(imageParser, consulInstallationFinder)

		consulContainer := consulDeployment().Spec.Template.Spec.Containers[0]
		deployment := &k8s_apps_v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "test-image",
							},
							consulContainer,
						},
					},
				},
			},
		}

		consulInstallationFinder.
			EXPECT().
			IsConsulConnect(deployment.Spec.Template.Spec.Containers[0]).
			Return(false, nil)

		consulInstallationFinder.
			EXPECT().
			IsConsulConnect(consulContainer).
			Return(true, nil)

		imageParser.
			EXPECT().
			Parse(consulContainer.Image).
			Return(&docker.Image{
				Domain: "test.com",
				Path:   "consul",
				Tag:    consulVersion,
			}, nil)

		expectedMesh := &mp_v1alpha1.Mesh{
			ObjectMeta: k8s_meta_v1.ObjectMeta{
				Name:      "consul-minidc-consul-ns",
				Namespace: env.DefaultWriteNamespace,
				Labels:    consul.DiscoveryLabels,
			},
			Spec: mp_v1alpha1_types.MeshSpec{
				MeshType: &mp_v1alpha1_types.MeshSpec_ConsulConnect{
					ConsulConnect: &mp_v1alpha1_types.ConsulConnectMesh{
						Installation: &mp_v1alpha1_types.MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               consulVersion,
						},
					},
				},
				Cluster: &mp_v1alpha1_types.ResourceRef{
					Name:      deployment.GetClusterName(),
					Namespace: env.DefaultWriteNamespace,
				},
			},
		}
		mesh, err := consulMeshFinder.ScanDeployment(ctx, deployment)

		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("reports the appropriate error when the installation finder errors", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)
		consulInstallationFinder := mock_consul.NewMockConsulConnectInstallationFinder(ctrl)

		consulMeshFinder := consul.NewConsulMeshFinder(imageParser, consulInstallationFinder)

		consulContainer := consulDeployment().Spec.Template.Spec.Containers[0]
		deployment := &k8s_apps_v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Image: "test-image",
							},
							consulContainer,
						},
					},
				},
			},
		}

		testErr := eris.New("test-err")

		consulInstallationFinder.
			EXPECT().
			IsConsulConnect(deployment.Spec.Template.Spec.Containers[0]).
			Return(false, nil)

		consulInstallationFinder.
			EXPECT().
			IsConsulConnect(consulContainer).
			Return(false, testErr)

		mesh, err := consulMeshFinder.ScanDeployment(ctx, deployment)

		Expect(mesh).To(BeNil())
		Expect(err).To(testutils.HaveInErrorChain(consul.ErrorDetectingDeployment(testErr)))
	})
})

func consulDeployment() *k8s_apps_v1.Deployment {
	return &k8s_apps_v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
		Spec: appsv1.DeploymentSpec{
			Template: kubev1.PodTemplateSpec{
				Spec: kubev1.PodSpec{
					Containers: []kubev1.Container{{
						Image: fmt.Sprintf("consul:%s", consulVersion),
						Command: []string{
							"/bin/sh",
							"-ec",
							`
CONSUL_FULLNAME="consul-consul"

exec /bin/consul agent \
  -advertise="${POD_IP}" \
  -bind=0.0.0.0 \
  -bootstrap-expect=1 \
  -client=0.0.0.0 \
  -config-dir=/consul/config \
  -datacenter=` + datacenterName + ` \
  -data-dir=/consul/data \
  -domain=consul \
  -hcl="connect { enabled = true }" \
  -ui \
  -retry-join=${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc \
  -server`,
						},
					}},
				},
			},
		},
	}
}
