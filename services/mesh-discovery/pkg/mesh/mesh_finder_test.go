package mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	mesh_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh"
	mock_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/mocks"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Mesh Finder", func() {
	var (
		ctrl            *gomock.Controller
		ctx             = context.TODO()
		clusterName     = "cluster-name"
		remoteNamespace = "remote-namespace"
		testErr         = eris.New("test-err")
		notFoundErr     = &errors.StatusError{
			ErrStatus: v1.Status{
				Reason: v1.StatusReasonNotFound,
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Create Event", func() {
		It("can discover a mesh", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := mock_discovery.BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Get(ctx, client.ObjectKey{
					Namespace: remoteNamespace,
					Name:      "test-mesh",
				}).
				Return(nil, notFoundErr)

			localMeshClient.
				EXPECT().
				Create(ctx, mesh).
				Return(nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("can go on to discover a mesh if one of the other finders errors out", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			brokenMeshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := mock_discovery.BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{brokenMeshFinder, meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Get(ctx, client.ObjectKey{
					Namespace: remoteNamespace,
					Name:      "test-mesh",
				}).
				Return(nil, notFoundErr)

			brokenMeshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
				Return(nil, testErr)

			localMeshClient.
				EXPECT().
				Create(ctx, mesh).
				Return(nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("responds with an error if no mesh was found and the finders reported an error", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
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
			deployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
				Return(nil, nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("doesn't write a CR if we discovered a mesh that we discovered previously", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := mock_discovery.BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Get(ctx, client.ObjectKey{
					Namespace: remoteNamespace,
					Name:      "test-mesh",
				}).
				Return(mesh, nil)

			err := eventHandler.Create(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})

		It("doesn't write a CR if we can't determine if this is a newly discovered mesh or not", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := mock_discovery.BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, deployment).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Get(ctx, client.ObjectKey{
					Namespace: remoteNamespace,
					Name:      "test-mesh",
				}).
				Return(nil, testErr)

			err := eventHandler.Create(deployment)
			Expect(err).To(Equal(testErr))
			Expect(deployment.GetClusterName()).To(Equal(clusterName))
		})
	})

	Context("Update Event", func() {
		It("can discover a mesh", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			oldDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "old-deployment", Namespace: remoteNamespace})
			newDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			meshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := mock_discovery.BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, oldDeployment).
				Return(nil, nil)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				Get(ctx, client.ObjectKey{
					Namespace: remoteNamespace,
					Name:      "test-mesh",
				}).
				Return(nil, notFoundErr)

			localMeshClient.
				EXPECT().
				Create(ctx, mesh).
				Return(nil)

			err := eventHandler.Update(oldDeployment, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(oldDeployment.GetClusterName()).To(Equal(clusterName))
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})

		It("does not do anything if we can't determine whether an old deployment represented a mesh", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			brokenMeshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			oldDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "old-deployment", Namespace: remoteNamespace})
			newDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{brokenMeshFinder, meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, oldDeployment).
				Return(nil, nil)

			brokenMeshFinder.
				EXPECT().
				ScanDeployment(ctx, oldDeployment).
				Return(nil, testErr)

			err := eventHandler.Update(oldDeployment, newDeployment)
			multierr, ok := err.(*multierror.Error)
			Expect(ok).To(BeTrue())
			Expect(multierr.Errors).To(HaveLen(1))
			Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(testErr))

			Expect(oldDeployment.GetClusterName()).To(Equal(clusterName))
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})

		It("doesn't do anything if no mesh was discovered", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			oldDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "old-deployment", Namespace: remoteNamespace})
			newDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, oldDeployment).
				Return(nil, nil)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment).
				Return(nil, nil)

			err := eventHandler.Update(oldDeployment, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(oldDeployment.GetClusterName()).To(Equal(clusterName))
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})

		It("doesn't write a CR if the discovered meshes are equal", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			oldDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "old-deployment", Namespace: remoteNamespace})
			newDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			meshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := mock_discovery.BuildMesh(meshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, oldDeployment).
				Return(mesh, nil)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment).
				Return(mesh, nil)

			err := eventHandler.Update(oldDeployment, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(oldDeployment.GetClusterName()).To(Equal(clusterName))
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})

		It("writes a new CR if an update event changes the mesh type", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			oldDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "old-deployment", Namespace: remoteNamespace})
			newDeployment := mock_discovery.BuildDeployment(v1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			oldMeshObjectMeta := v1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			oldMesh := mock_discovery.BuildMesh(oldMeshObjectMeta)

			newMeshObjectMeta := v1.ObjectMeta{Name: "new-test-mesh", Namespace: remoteNamespace}
			newMesh := mock_discovery.BuildMesh(newMeshObjectMeta)

			eventHandler := mesh_discovery.DefaultMeshFinder(
				ctx,
				clusterName,
				[]mesh_discovery.MeshScanner{meshFinder},
				localMeshClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, oldDeployment).
				Return(oldMesh, nil)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, newDeployment).
				Return(newMesh, nil)

			localMeshClient.
				EXPECT().
				Get(ctx, client.ObjectKey{
					Namespace: remoteNamespace,
					Name:      "new-test-mesh",
				}).
				Return(nil, notFoundErr)

			localMeshClient.
				EXPECT().
				Create(ctx, newMesh).
				Return(nil)

			err := eventHandler.Update(oldDeployment, newDeployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(oldDeployment.GetClusterName()).To(Equal(clusterName))
			Expect(newDeployment.GetClusterName()).To(Equal(clusterName))
		})
	})
})
