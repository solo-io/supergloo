package discovery_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery"
	mock_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mocks"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Local mesh client", func() {
	var (
		ctrl *gomock.Controller
		ctx  = context.TODO()
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("forwards on operations to the dynamic client", func() {
		mgr := mock_mc_manager.NewMockAsyncManager(ctrl)
		k8s_mgr := mock_controller_runtime.NewMockManager(ctrl)
		client := mock_controller_runtime.NewMockClient(ctrl)

		meshObjectMeta := controllerruntime.ObjectMeta{
			Name:      "test-mesh",
			Namespace: "ns",
		}
		mesh := mock_discovery.BuildMesh(meshObjectMeta)

		mgr.
			EXPECT().
			Manager().
			Return(k8s_mgr)

		k8s_mgr.
			EXPECT().
			GetClient().
			Return(client)

		client.EXPECT().Create(ctx, mesh).Return(nil)
		client.
			EXPECT().
			Get(ctx, client2.ObjectKey{
				Name:      meshObjectMeta.Name,
				Namespace: meshObjectMeta.Namespace,
			}, gomock.Any()).
			DoAndReturn(func(ctx context.Context, key client2.ObjectKey, obj runtime.Object) error {
				sideEffectedMesh, ok := obj.(*v1alpha1.Mesh)
				Expect(ok).To(BeTrue())

				sideEffectedMesh.Name = meshObjectMeta.Name
				sideEffectedMesh.Namespace = meshObjectMeta.Namespace

				return nil
			})

		client.EXPECT().Delete(ctx, mesh).Return(nil)

		localMeshClient := discovery.NewLocalMeshClient(mgr)
		Expect(localMeshClient.Create(ctx, mesh)).To(BeNil())

		returnedMesh, err := localMeshClient.Get(ctx, client2.ObjectKey{
			Name:      meshObjectMeta.Name,
			Namespace: meshObjectMeta.Namespace,
		})
		Expect(err).To(BeNil())
		Expect(returnedMesh).To(Equal(mesh))

		Expect(localMeshClient.Delete(ctx, mesh)).To(BeNil())
	})

	It("does not return an error from get if the object is not found", func() {
		mgr := mock_mc_manager.NewMockAsyncManager(ctrl)
		k8s_mgr := mock_controller_runtime.NewMockManager(ctrl)
		client := mock_controller_runtime.NewMockClient(ctrl)

		meshObjectMeta := controllerruntime.ObjectMeta{
			Name:      "test-mesh",
			Namespace: "ns",
		}

		mgr.
			EXPECT().
			Manager().
			Return(k8s_mgr)

		k8s_mgr.
			EXPECT().
			GetClient().
			Return(client)

		k8sNotFoundErr := &errors.StatusError{
			ErrStatus: v1.Status{
				Reason: v1.StatusReasonNotFound,
			},
		}
		client.
			EXPECT().
			Get(ctx, client2.ObjectKey{
				Name:      meshObjectMeta.Name,
				Namespace: meshObjectMeta.Namespace,
			}, gomock.Any()).
			Return(k8sNotFoundErr)

		localMeshClient := discovery.NewLocalMeshClient(mgr)
		returnedMesh, err := localMeshClient.Get(ctx, client2.ObjectKey{
			Name:      meshObjectMeta.Name,
			Namespace: meshObjectMeta.Namespace,
		})

		Expect(err).To(Equal(k8sNotFoundErr)) // this sould be strict equality
		Expect(returnedMesh).To(BeNil())
	})
})
