package consul_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	mock_v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker"
	mock_docker "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/consul"
	mock_consul "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/consul/mocks"
	k8s_apps "k8s.io/api/apps/v1"
	k8s_core "k8s.io/api/core/v1"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	consulNs       = "consul-ns"
	consulVersion  = "1.6.2"
	datacenterName = "minidc"
)

var _ = Describe("Consul Mesh Finder", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		mockConfigMapClient = mock_v1.NewMockConfigMapClient(ctrl)
		clusterName         = "test-cluster-name"
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
		consulInstallationFinder := mock_consul.NewMockConsulConnectInstallationScanner(ctrl)

		consulMeshFinder := consul.NewConsulMeshScanner(imageParser, consulInstallationFinder)

		nonConsulDeployment := &k8s_apps.Deployment{
			ObjectMeta: k8s_meta.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps.DeploymentSpec{
				Template: k8s_core.PodTemplateSpec{
					Spec: k8s_core.PodSpec{
						Containers: []k8s_core.Container{{
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

		mesh, err := consulMeshFinder.ScanDeployment(ctx, clusterName, nonConsulDeployment, mockConfigMapClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("can discover consul", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)
		consulInstallationFinder := mock_consul.NewMockConsulConnectInstallationScanner(ctrl)

		consulMeshFinder := consul.NewConsulMeshScanner(imageParser, consulInstallationFinder)

		consulContainer := consulDeployment().Spec.Template.Spec.Containers[0]
		deployment := &k8s_apps.Deployment{
			ObjectMeta: k8s_meta.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps.DeploymentSpec{
				Template: k8s_core.PodTemplateSpec{
					Spec: k8s_core.PodSpec{
						Containers: []k8s_core.Container{
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

		expectedMesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta.ObjectMeta{
				Name:      "consul-minidc-consul-ns",
				Namespace: container_runtime.GetWriteNamespace(),
				Labels:    consul.DiscoveryLabels,
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_ConsulConnect{
					ConsulConnect: &smh_discovery_types.MeshSpec_ConsulConnectMesh{
						Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               consulVersion,
						},
					},
				},
				Cluster: &smh_core_types.ResourceRef{
					Name:      clusterName,
					Namespace: container_runtime.GetWriteNamespace(),
				},
			},
		}
		mesh, err := consulMeshFinder.ScanDeployment(ctx, clusterName, deployment, mockConfigMapClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("reports the appropriate error when the installation finder errors", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)
		consulInstallationFinder := mock_consul.NewMockConsulConnectInstallationScanner(ctrl)

		consulMeshFinder := consul.NewConsulMeshScanner(imageParser, consulInstallationFinder)

		consulContainer := consulDeployment().Spec.Template.Spec.Containers[0]
		deployment := &k8s_apps.Deployment{
			ObjectMeta: k8s_meta.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps.DeploymentSpec{
				Template: k8s_core.PodTemplateSpec{
					Spec: k8s_core.PodSpec{
						Containers: []k8s_core.Container{
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

		mesh, err := consulMeshFinder.ScanDeployment(ctx, clusterName, deployment, mockConfigMapClient)

		Expect(mesh).To(BeNil())
		Expect(err).To(testutils.HaveInErrorChain(consul.ErrorDetectingDeployment(testErr)))
	})
})

func consulDeployment() *k8s_apps.Deployment {
	return &k8s_apps.Deployment{
		ObjectMeta: k8s_meta.ObjectMeta{Namespace: consulNs, Name: "name doesn't matter in this context"},
		Spec: k8s_apps.DeploymentSpec{
			Template: k8s_core.PodTemplateSpec{
				Spec: k8s_core.PodSpec{
					Containers: []k8s_core.Container{{
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
