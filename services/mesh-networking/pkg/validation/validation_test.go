package vm_validation_test

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/snapshot"
	vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation"
	mock_vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validator", func() {

	var (
		ctrl              *gomock.Controller
		validator         snapshot.MeshNetworkingSnapshotValidator
		meshFinder        *mock_vm_validation.MockVirtualMeshFinder
		virtualMeshClient *mock_zephyr_networking.MockVirtualMeshClient
		ctx               context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		meshFinder = mock_vm_validation.NewMockVirtualMeshFinder(ctrl)
		virtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		validator = vm_validation.NewVirtualMeshValidator(meshFinder, virtualMeshClient)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will return invalid if a mesh ref doesn't exist", func() {
		ref := &zephyr_core_types.ResourceRef{
			Name:      "incorrect",
			Namespace: "ref",
		}
		vm := &zephyr_networking.VirtualMesh{
			Spec: zephyr_networking_types.VirtualMeshSpec{
				Meshes: []*zephyr_core_types.ResourceRef{ref},
			},
		}
		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return(nil, testErr)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeFalse())
		Expect(vm.Status.ConfigStatus.State).To(Equal(zephyr_core_types.Status_INVALID))
	})

	It("will recover to valid after being invalid", func() {
		ref := &zephyr_core_types.ResourceRef{
			Name:      "incorrect",
			Namespace: "ref",
		}
		vm := &zephyr_networking.VirtualMesh{
			Spec: zephyr_networking_types.VirtualMeshSpec{
				Meshes: []*zephyr_core_types.ResourceRef{ref},
			},
			Status: zephyr_networking_types.VirtualMeshStatus{
				ConfigStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_INVALID,
					Message: testErr.Error(),
				},
			},
		}

		// A valid mesh exists
		mesh := zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{
					Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
				},
			},
		}
		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*zephyr_discovery.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeTrue())


	})

	It("will return invalid if a non-istio mesh is referenced", func() {
		ref := &zephyr_core_types.ResourceRef{
			Name:      "valid",
			Namespace: "ref",
		}
		mesh := zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_ConsulConnect{},
			},
		}
		vm := &zephyr_networking.VirtualMesh{
			Spec: zephyr_networking_types.VirtualMeshSpec{
				Meshes: []*zephyr_core_types.ResourceRef{ref},
			},
			Status: zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_INVALID,
					Message: vm_validation.OnlyIstioSupportedError(mesh.Name).Error(),
				},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*zephyr_discovery.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeFalse())
	})

	It("will return valid and no error if all went fine", func() {
		ref := &zephyr_core_types.ResourceRef{
			Name:      "valid",
			Namespace: "ref",
		}
		mesh := zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{
					Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
				},
			},
		}
		vm := &zephyr_networking.VirtualMesh{
			Spec: zephyr_networking_types.VirtualMeshSpec{
				Meshes: []*zephyr_core_types.ResourceRef{ref},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*zephyr_discovery.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeTrue())
	})
})
