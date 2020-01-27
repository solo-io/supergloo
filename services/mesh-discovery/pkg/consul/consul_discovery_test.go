package consul_test

import (
	"context"
	"fmt"

	mock_docker "github.com/solo-io/mesh-projects/pkg/common/docker/mocks"

	"github.com/solo-io/mesh-projects/pkg/common/docker"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	zeph_core "github.com/solo-io/mesh-projects/pkg/api/v1/core"
	mocks_common "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/consul"
	mock_consul "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/consul/mocks"
	"github.com/solo-io/solo-kit/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	writeNs        = "write-ns"
	consulNs       = "consul-ns"
	consulVersion  = "1.6.2"
	datacenterName = "minidc"
)

var _ = Describe("Consul Discovery", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("reconciles nil", func() {
		snap := &v1.DiscoverySnapshot{}

		syncer, reconciler, _, _, _ := buildGlobalObjects(ctrl)
		err := syncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		Expect(reconciler.ReconcileCalledWith).To(HaveLen(1))
		Expect(reconciler.ReconcileCalledWith[0]).To(HaveLen(0))
	})

	It("can discover a simple deployment", func() {
		expectedMeshName := "consul-minidc-consul-ns"
		expectedMesh := &v1.Mesh{
			Metadata: core.Metadata{
				Name:      expectedMeshName,
				Namespace: writeNs,
				Labels:    consul.DiscoveryLabels,
			},
			MeshType: &v1.Mesh_ConsulConnect{
				ConsulConnect: &v1.ConsulConnectMesh{
					Installation: &v1.MeshInstallation{
						InstallationNamespace: consulNs,
						Version:               consulVersion,
					},
				},
			},
			EntryPoint: &zeph_core.ClusterResourceRef{
				Resource: core.ResourceRef{
					Name:      expectedMeshName,
					Namespace: writeNs,
				},
			},
		}
		snap := &v1.DiscoverySnapshot{Deployments: kubernetes.DeploymentList{consulDeployment()}}

		syncer, reconciler, _, consulConnectFinder, imageNameParser := buildGlobalObjects(ctrl)

		consulContainer := snap.Deployments[0].Spec.Template.Spec.Containers[0]
		imageNameParser.EXPECT().
			Parse(consulContainer.Image).
			Return(&docker.Image{
				Domain: "docker.io",
				Path:   "library/consul",
				Tag:    "1.6.2",
			}, nil)
		consulConnectFinder.EXPECT().
			IsConsulConnect(consulContainer).
			Return(true, nil)

		err := syncer.Sync(context.TODO(), snap)

		Expect(err).NotTo(HaveOccurred())
		Expect(reconciler.ReconcileCalledWith).To(HaveLen(1))
		Expect(reconciler.ReconcileCalledWith[0]).To(HaveLen(1))
		Expect(reconciler.ReconcileCalledWith[0][0]).To(Equal(expectedMesh))
	})
})

func buildGlobalObjects(ctrl *gomock.Controller) (v1.DiscoverySyncer,
	*mocks_common.MockMeshReconciler,
	*mocks_common.MockMeshIngressReconciler,
	*mock_consul.MockConsulConnectInstallationFinder,
	*mock_docker.MockImageNameParser) {

	reconciler := &mocks_common.MockMeshReconciler{}
	ingressReconciler := &mocks_common.MockMeshIngressReconciler{}
	mockConnectDeploymentFinder := mock_consul.NewMockConsulConnectInstallationFinder(ctrl)
	mockImageNameParser := mock_docker.NewMockImageNameParser(ctrl)

	return consul.NewConsulDiscoveryPlugin(writeNs, reconciler, ingressReconciler, mockConnectDeploymentFinder, mockImageNameParser),
		reconciler,
		ingressReconciler,
		mockConnectDeploymentFinder,
		mockImageNameParser
}

func consulDeployment() *kubernetes.Deployment {
	return &kubernetes.Deployment{
		Deployment: deployment.Deployment{
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
		},
	}
}
