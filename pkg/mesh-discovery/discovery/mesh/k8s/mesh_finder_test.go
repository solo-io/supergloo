package k8s_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	mp_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/kube"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	mock_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_k8s_apps_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/apps/v1"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				UpsertMeshSpec(ctx, mesh).
				Return(nil)

			err := eventHandler.CreateDeployment(deployment)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can go on to discover a mesh if one of the other finders errors out", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			brokenMeshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{brokenMeshFinder, meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				UpsertMeshSpec(ctx, mesh).
				Return(nil)

			brokenMeshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(nil, testErr)

			err := eventHandler.CreateDeployment(deployment)
			Expect(err).NotTo(HaveOccurred())

		})

		It("responds with an error if no mesh was found and the finders reported an error", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(nil, testErr)

			err := eventHandler.CreateDeployment(deployment)
			multierr, ok := err.(*multierror.Error)
			Expect(ok).To(BeTrue())
			Expect(multierr.Errors).To(HaveLen(1))
			Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(testErr))

		})

		It("doesn't do anything if no mesh was discovered", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(nil, nil)

			err := eventHandler.CreateDeployment(deployment)
			Expect(err).NotTo(HaveOccurred())

		})

		It("performs an upsert if we discovered a mesh that we discovered previously", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				UpsertMeshSpec(ctx, mesh).
				Return(nil)

			err := eventHandler.CreateDeployment(deployment)
			Expect(err).NotTo(HaveOccurred())

		})

		It("returns error from Upsert if upsert fails", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				UpsertMeshSpec(ctx, mesh).
				Return(testErr)

			err := eventHandler.CreateDeployment(deployment)
			Expect(err).To(Equal(testErr))

		})
	})

	Context("Update Event", func() {
		It("can discover a mesh", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			newDeployment := BuildDeployment(metav1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, newDeployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				UpsertMeshSpec(ctx, mesh).
				Return(nil)

			err := eventHandler.UpdateDeployment(nil, newDeployment)
			Expect(err).NotTo(HaveOccurred())

		})

		It("doesn't do anything if no mesh was discovered", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			newDeployment := BuildDeployment(metav1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, newDeployment, clusterClient).
				Return(nil, nil)

			err := eventHandler.UpdateDeployment(nil, newDeployment)
			Expect(err).NotTo(HaveOccurred())

		})

		It("writes a new CR if an update event changes the mesh type", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			newDeployment := BuildDeployment(metav1.ObjectMeta{Name: "new-deployment", Namespace: remoteNamespace})
			newMeshObjectMeta := metav1.ObjectMeta{Name: "new-test-mesh", Namespace: remoteNamespace}
			newMesh := BuildMesh(newMeshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)
			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, newDeployment, clusterClient).
				Return(newMesh, nil)
			localMeshClient.
				EXPECT().
				UpsertMeshSpec(ctx, newMesh).
				Return(nil)
			err := eventHandler.UpdateDeployment(nil, newDeployment)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Context("Delete Event", func() {
		It("can delete a mesh when appropriate", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(mesh, nil)

			localMeshClient.
				EXPECT().
				DeleteMesh(ctx, selection.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(nil)

			err := eventHandler.DeleteDeployment(deployment)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not delete a mesh if the deployment is not a control plane", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(nil, nil)

			err := eventHandler.DeleteDeployment(deployment)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Startup Reconciliation", func() {
		It("does nothing if there is nothing discovered yet", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)

			localMeshClient.EXPECT().
				ListMesh(ctx, client.MatchingLabels{kube.COMPUTE_TARGET: clusterName}).
				Return(&mp_v1alpha1.MeshList{Items: []mp_v1alpha1.Mesh{}}, nil)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			err := eventHandler.StartDiscovery(noOpDeploymentEventWatcher{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("does nothing if the state is up-to-date", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			localMeshClient.EXPECT().
				ListMesh(ctx, client.MatchingLabels{kube.COMPUTE_TARGET: clusterName}).
				Return(&mp_v1alpha1.MeshList{Items: []mp_v1alpha1.Mesh{*mesh}}, nil)

			deploymentClient.EXPECT().
				ListDeployment(ctx).
				Return(&appsv1.DeploymentList{Items: []appsv1.Deployment{*deployment}}, nil)

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(mesh, nil)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			err := eventHandler.StartDiscovery(noOpDeploymentEventWatcher{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes meshes that no longer exist", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)

			localMeshClient.EXPECT().
				ListMesh(ctx, client.MatchingLabels{kube.COMPUTE_TARGET: clusterName}).
				Return(&mp_v1alpha1.MeshList{Items: []mp_v1alpha1.Mesh{*mesh}}, nil)

			deploymentClient.EXPECT().
				ListDeployment(ctx).
				Return(&appsv1.DeploymentList{Items: []appsv1.Deployment{}}, nil)

			localMeshClient.EXPECT().
				DeleteMesh(ctx, selection.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(nil)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			err := eventHandler.StartDiscovery(noOpDeploymentEventWatcher{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates meshes that are out-of-date with the current state", func() {
			meshFinder := mock_discovery.NewMockMeshScanner(ctrl)
			localMeshClient := mock_core.NewMockMeshClient(ctrl)
			deploymentClient := mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
			deployment := BuildDeployment(metav1.ObjectMeta{Name: "test-deployment", Namespace: remoteNamespace})
			meshObjectMeta := metav1.ObjectMeta{Name: "test-mesh", Namespace: remoteNamespace}
			mesh := BuildMesh(meshObjectMeta)
			mesh.Spec = types.MeshSpec{
				MeshType: &types.MeshSpec_Istio1_5_{
					Istio1_5: &types.MeshSpec_Istio1_5{
						Metadata: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								Version: "1.5.0",
							},
						},
					},
				},
			}

			localMeshClient.EXPECT().
				ListMesh(ctx, client.MatchingLabels{kube.COMPUTE_TARGET: clusterName}).
				Return(&mp_v1alpha1.MeshList{Items: []mp_v1alpha1.Mesh{*mesh}}, nil)

			deploymentClient.EXPECT().
				ListDeployment(ctx).
				Return(&appsv1.DeploymentList{Items: []appsv1.Deployment{*deployment}}, nil)

			updatedMesh := *mesh
			updatedMesh.Spec = types.MeshSpec{
				MeshType: &types.MeshSpec_Istio1_5_{
					Istio1_5: &types.MeshSpec_Istio1_5{
						Metadata: &types.MeshSpec_IstioMesh{
							Installation: &types.MeshSpec_MeshInstallation{
								Version: "1.5.1",
							},
						},
					},
				},
			}

			meshFinder.
				EXPECT().
				ScanDeployment(ctx, clusterName, deployment, clusterClient).
				Return(&updatedMesh, nil)

			localMeshClient.EXPECT().
				UpdateMesh(ctx, &updatedMesh).
				Return(nil)

			eventHandler := k8s.NewMeshFinder(
				ctx,
				clusterName,
				[]k8s.MeshScanner{meshFinder},
				localMeshClient,
				clusterClient,
				deploymentClient,
			)

			err := eventHandler.StartDiscovery(noOpDeploymentEventWatcher{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

type noOpDeploymentEventWatcher struct{}

var _ controller.DeploymentEventWatcher = noOpDeploymentEventWatcher{}

func (noOpDeploymentEventWatcher) AddEventHandler(ctx context.Context, h controller.DeploymentEventHandler, predicates ...predicate.Predicate) error {
	return nil
}
