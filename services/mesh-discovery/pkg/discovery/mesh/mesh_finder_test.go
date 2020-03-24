package mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	mp_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	mesh_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh"
	mock_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh/mocks"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildDeployment(objMeta metav1.ObjectMeta) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: objMeta,
	}
}

func BuildMesh(objMeta metav1.ObjectMeta) *mp_v1alpha1.Mesh {
	return &mp_v1alpha1.Mesh{
		ObjectMeta: objMeta,
	}
}

var _ = Describe("Mesh Finder", func() {
	var (
		ctrl            *gomock.Controller
		ctx             = context.TODO()
		clusterName     = "cluster-name"
		remoteNamespace = "remote-namespace"
		clusterClient   *mock_controller_runtime.MockClient
		testErr         = eris.New("test-err")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		clusterClient = mock_controller_runtime.NewMockClient(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Create Event", func() {
		It("can discover a mesh", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Upsert(ctx, mesh).
				Return(nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("can go on to discover a mesh if one of the other finders errors out", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			brokenMeshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{brokenMeshFinder, meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Upsert(ctx, mesh).
				Return(nil)

			brokenMeshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(nil, testErr)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("responds with an error if no mesh was found and the finders reported an error", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(nil, testErr)

			err := eventHandler.Create(deployment)
			multierr, ok := err.(*multierror.Error)
			Expect(ok).To(BeTrue())
			Expect(multierr.Errors).To(HaveLen(1))
			Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(testErr))
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("doesn't do anything if no mesh was discovered", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(nil, nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("performs an upsert if we discovered a mesh that we discovered previously", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Upsert(ctx, mesh).
				Return(nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("returns error from UpsertData if upsert fails", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Upsert(ctx, mesh).
				Return(testErr)

			err := eventHandler.Create(deployment)
			Expect(err).To(Equal(testErr))
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})
	})

	Context("Update Event", func() {
		It("can discover a mesh", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			newDeployment := BuildDeployment(metav1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Upsert(ctx, mesh).
				Return(nil)

			err := eventHandler.Update(nil, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})

		It("doesn't do anything if no mesh was discovered", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			newDeployment := BuildDeployment(metav1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment, clusterClient).
				Return(nil, nil)

			err := eventHandler.Update(nil, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})

		It("writes a new CR if an update event changes the mesh type", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			newDeployment := BuildDeployment(metav1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			newMeshObjectMeta := metav1.ObjectMeta{Name: "new-test-mesh", Namespace: remoteNamespace}
			newMesh := BuildMesh(newMeshObjectMeta)

			eventHandler := mesh_discovery.NewMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
			)
			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment, clusterClient).
				Return(newMesh, nil)
			localMeshClient.
				EXPECT().
				Upsert(ctx, newMesh).
				Return(nil)
			err := eventHandler.Update(nil, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})
	})
})
