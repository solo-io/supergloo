package vm_validation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networkingv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	v1alpha1_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/snapshot"
	vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation"
	mock_vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		ref := &core_types.ResourceRef{
			Name:      "incorrect",
			Namespace: "ref",
		}
		vm := &networkingv1alpha1.VirtualMesh{
			Spec: v1alpha1_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{ref},
			},
			Status: v1alpha1_types.VirtualMeshStatus{
				CertificateStatus: &core_types.Status{
					State:   core_types.Status_INVALID,
					Message: testErr.Error(),
				},
			},
		}
		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return(nil, testErr)

		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeFalse())
	})

	It("will return invalid if a non-istio mesh is referenced", func() {
		ref := &core_types.ResourceRef{
			Name:      "valid",
			Namespace: "ref",
		}
		mesh := discoveryv1alpha1.Mesh{
			ObjectMeta: v1.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_ConsulConnect{},
			},
		}
		vm := &networkingv1alpha1.VirtualMesh{
			Spec: v1alpha1_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{ref},
			},
			Status: v1alpha1_types.VirtualMeshStatus{
				CertificateStatus: &core_types.Status{
					State:   core_types.Status_INVALID,
					Message: vm_validation.OnlyIstioSupportedError(mesh.Name).Error(),
				},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*discoveryv1alpha1.Mesh{&mesh}, nil)

		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, vm).
			Return(nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeFalse())
	})

	It("will return valid and no error if all went fine", func() {
		ref := &core_types.ResourceRef{
			Name:      "valid",
			Namespace: "ref",
		}
		mesh := discoveryv1alpha1.Mesh{
			ObjectMeta: v1.ObjectMeta{
				Name:      ref.GetName(),
				Namespace: ref.GetNamespace(),
			},
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{},
				},
			},
		}
		vm := &networkingv1alpha1.VirtualMesh{
			Spec: v1alpha1_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{ref},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForVirtualMesh(ctx, vm).
			Return([]*discoveryv1alpha1.Mesh{&mesh}, nil)

		valid := validator.ValidateVirtualMeshUpsert(ctx, vm, nil)
		Expect(valid).To(BeTrue())
	})
})
