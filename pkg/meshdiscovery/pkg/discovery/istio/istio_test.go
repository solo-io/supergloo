package istio

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("istio mesh discovery unit tests", func() {
	var (
		snap    *v1.DiscoverySnapshot
		enabled *common.EnabledConfigLoops
		syncer  *istioMeshDiscovery
		ctx     context.Context

		istioNamespace     = "istio-system"
		superglooNamespace = "supergloo-system"
	)

	var constructPod = func(container kubev1.Container, namespace string) *v1.Pod {

		pod := &kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "istio-pilot",
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					container,
				},
			},
		}
		return kubernetes.FromKube(pod)
	}

	BeforeEach(func() {
		snap = &v1.DiscoverySnapshot{
			Meshes: v1.MeshesByNamespace{
				superglooNamespace: v1.MeshList{
					{
						Metadata: core.Metadata{Name: "one"},
						MeshType: &v1.Mesh_Istio{
							Istio: &v1.IstioMesh{},
						},
					},
					{
						MeshType: &v1.Mesh_AwsAppMesh{
							AwsAppMesh: &v1.AwsAppMesh{},
						},
					},
				},
			},
		}
		enabled = &common.EnabledConfigLoops{}
		syncer = NewIstioMeshDiscovery()
		ctx = context.TODO()
		clients.UseMemoryClients()
	})
	Context("properly sets istio enabled", func() {
		It("sets false properly with no istio data discovered", func() {
			err := syncer.DiscoverMeshes(ctx, &v1.DiscoverySnapshot{}, enabled)
			Expect(err).NotTo(HaveOccurred())
			Expect(enabled.Istio()).To(BeFalse())
		})
		It("sets true properly with existing mesh", func() {
			container := kubev1.Container{
				Image: "istio-",
			}
			pod := constructPod(container, istioNamespace)
			snap = &v1.DiscoverySnapshot{}
			snap.Pods = v1.PodsByNamespace{
				istioNamespace: v1.PodList{
					pod,
				},
			}
			err := syncer.DiscoverMeshes(ctx, snap, enabled)
			Expect(err).NotTo(HaveOccurred())
			Expect(enabled.Istio()).To(BeTrue())
		})
		It("sets true properly with pod", func() {
			container := kubev1.Container{
				Image: "istio-",
			}
			pod := constructPod(container, istioNamespace)
			snap.Pods = v1.PodsByNamespace{
				istioNamespace: v1.PodList{
					pod,
				},
			}
			err := syncer.DiscoverMeshes(ctx, snap, enabled)
			Expect(err).NotTo(HaveOccurred())
			Expect(enabled.Istio()).To(BeTrue())
		})
	})
})
