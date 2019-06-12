package appmesh

//
//import (
//	"context"
//
//	"github.com/golang/mock/gomock"
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
//	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
//	v1 "github.com/solo-io/supergloo/pkg/api/v1"
//	"github.com/solo-io/supergloo/pkg/api/v1/mocks"
//	appmesh2 "github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/appmesh"
//	"github.com/solo-io/supergloo/test/inputs/appmesh"
//	"github.com/solo-io/supergloo/test/inputs/appmesh/scenarios"
//)
//
//var _ = Describe("appmesh config syncer", func() {
//
//	var (
//		ctx        context.Context
//		syncer     *appmeshDiscoveryConfigSyncer
//		input      *appmesh.PodsServicesUpstreamsTuple
//		ctrl       *gomock.Controller
//		reconciler *mocks.MockMeshReconciler
//	)
//
//	var createMesh = func(name, namespace string, selectors map[string]string) *v1.Mesh {
//		mesh := &v1.Mesh{
//			Metadata: core.Metadata{
//				Labels:    selectors,
//				Name:      name,
//				Namespace: namespace,
//			},
//			MeshType: &v1.Mesh_AwsAppMesh{
//				AwsAppMesh: &v1.AwsAppMesh{},
//			},
//		}
//		return mesh
//	}
//
//	BeforeEach(func() {
//		ctrl = gomock.NewController(T)
//		ctx = context.TODO()
//		reconciler = mocks.NewMockMeshReconciler(ctrl)
//		syncer = newAppmeshDiscoveryConfigSyncer(reconciler)
//		input = scenarios.SumAppMeshRelatedResources()
//	})
//	AfterEach(func() {
//		ctrl.Finish()
//	})
//
//	It("returns nil with 0 meshes", func() {
//		snap := &v1.AppmeshDiscoverySnapshot{}
//		reconciler.EXPECT().Reconcile(gomock.Any(), nil, gomock.Any(), gomock.Any()).Times(1)
//		Expect(syncer.Sync(ctx, snap)).NotTo(HaveOccurred())
//	})
//
//	It("will only pick up meshes with correct labels", func() {
//		snap := &v1.AppmeshDiscoverySnapshot{
//			Meshes: v1.MeshList{
//				createMesh("one", "one", appmesh2.DiscoverySelector),
//				createMesh("two", "one", nil),
//			},
//		}
//		reconciler.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
//			DoAndReturn(func(namespace string, desiredResources v1.MeshList, transition v1.TransitionMeshFunc, opts clients.ListOpts) error {
//				Expect(desiredResources).To(HaveLen(1))
//				Expect(desiredResources[0].Metadata).To(BeEquivalentTo(core.Metadata{
//					Name:      "one",
//					Namespace: "one",
//					Labels:    appmesh2.DiscoverySelector,
//				}))
//				return nil
//			})
//		err := syncer.Sync(ctx, snap)
//		Expect(err).NotTo(HaveOccurred())
//	})
//
//	It("works with multiple meshes", func() {
//		snap := &v1.AppmeshDiscoverySnapshot{
//			Pods:      input.MustGetPodList(),
//			Upstreams: input.MustGetUpstreamList(),
//			Meshes: v1.MeshList{
//				createMesh("one", "one", appmesh2.DiscoverySelector),
//				createMesh(scenarios.MeshName, "one", appmesh2.DiscoverySelector),
//			},
//		}
//		reconciler.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
//			DoAndReturn(func(namespace string, desiredResources v1.MeshList, transition v1.TransitionMeshFunc, opts clients.ListOpts) error {
//				Expect(desiredResources).To(HaveLen(2))
//				for _, mesh := range desiredResources {
//					if mesh.Metadata.Name == scenarios.MeshName {
//						Expect(mesh.DiscoveryMetadata.Upstreams).To(HaveLen(9))
//
//					} else {
//						Expect(mesh.DiscoveryMetadata.Upstreams).To(BeNil())
//					}
//				}
//				return nil
//			})
//		err := syncer.Sync(ctx, snap)
//		Expect(err).NotTo(HaveOccurred())
//	})
//
//})
