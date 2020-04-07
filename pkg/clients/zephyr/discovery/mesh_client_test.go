package zephyr_discovery_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildDeployment(objMeta metav1.ObjectMeta) *k8s_apps_v1.Deployment {
	return &k8s_apps_v1.Deployment{
		ObjectMeta: objMeta,
	}
}

func BuildMesh(objMeta metav1.ObjectMeta) *v1alpha1.Mesh {
	return &v1alpha1.Mesh{
		ObjectMeta: objMeta,
	}
}

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
		client := mock_controller_runtime.NewMockClient(ctrl)

		meshObjectMeta := controllerruntime.ObjectMeta{
			Name:      "test-mesh",
			Namespace: "ns",
		}
		mesh := BuildMesh(meshObjectMeta)

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

		localMeshClient := zephyr_core.NewMeshClient(client)
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
		client := mock_controller_runtime.NewMockClient(ctrl)

		meshObjectMeta := controllerruntime.ObjectMeta{
			Name:      "test-mesh",
			Namespace: "ns",
		}

		k8sNotFoundErr := &errors.StatusError{
			ErrStatus: metav1.Status{
				Reason: metav1.StatusReasonNotFound,
			},
		}
		client.
			EXPECT().
			Get(ctx, client2.ObjectKey{
				Name:      meshObjectMeta.Name,
				Namespace: meshObjectMeta.Namespace,
			}, gomock.Any()).
			Return(k8sNotFoundErr)

		localMeshClient := zephyr_core.NewMeshClient(client)
		returnedMesh, err := localMeshClient.Get(ctx, client2.ObjectKey{
			Name:      meshObjectMeta.Name,
			Namespace: meshObjectMeta.Namespace,
		})

		Expect(err).To(Equal(k8sNotFoundErr)) // this sould be strict equality
		Expect(returnedMesh).To(BeNil())
	})
})
