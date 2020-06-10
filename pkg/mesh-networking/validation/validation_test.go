package vm_validation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	networking_snapshot "github.com/solo-io/service-mesh-hub/pkg/common/networking-snapshot"
	vm_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/validation"
	mock_vm_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/validation/mocks"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validator", func() {

	var (
		ctrl              *gomock.Controller
		validator         networking_snapshot.MeshNetworkingSnapshotValidator
		meshFinder        *mock_vm_validation.MockVirtualMeshFinder
		virtualMeshClient *mock_smh_networking.MockVirtualMeshClient
		ctx               context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		meshFinder = mock_vm_validation.NewMockVirtualMeshFinder(ctrl)
		virtualMeshClient = mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		validator = vm_validation.NewVirtualMeshValidator(meshFinder, virtualMeshClient)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will return invalid if a mesh ref doesn't exist", func() {
		ref := &smh_core_types.ResourceRef{
			Name:      "incorrect",
			Namespace: "ref",
		}
		vm := &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{ref},
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
		Expect(vm.Status.ConfigStatus.State).To(Equal(smh_core_types.Status_INVALID))
	})

	It("will recover to valid after being invalid", func() {
		ref := &smh_core_types.ResourceRef{
			Name:      "incorrect",
			Namespace: "ref",
		}
		vm := &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{ref},
			},
			Status: smh_networking_types.VirtualMeshStatus{
				ConfigStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: testErr.Error(),
				},
			},
		}

		// A valid mesh exists
		mesh := smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
				},
			},
		}
		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*smh_discovery.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeTrue())

	})

	It("will return invalid if a non-istio mesh is referenced", func() {
		ref := &smh_core_types.ResourceRef{
			Name:      "valid",
			Namespace: "ref",
		}
		mesh := smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_ConsulConnect{},
			},
		}
		vm := &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{ref},
			},
			Status: smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: vm_validation.MeshTypeNotSupportedError(mesh.Name).Error(),
				},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*smh_discovery.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeFalse())
	})

	It("will return valid and no error if all went fine", func() {
		ref := &smh_core_types.ResourceRef{
			Name:      "valid",
			Namespace: "ref",
		}
		mesh := smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
				},
			},
		}
		vm := &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{ref},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*smh_discovery.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeTrue())
	})
})
