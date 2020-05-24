package access_policy_enforcer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	global_ac_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer"
	mock_global_access_control_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking2 "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/networking"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("EnforcerLoop", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		virtualMeshEventWatcher *mock_zephyr_networking2.MockVirtualMeshEventWatcher
		virtualMeshClient       *mock_zephyr_networking.MockVirtualMeshClient
		meshClient              *mock_core.MockMeshClient
		meshEnforcers           []*mock_global_access_control_enforcer.MockAccessPolicyMeshEnforcer
		enforcerLoop            global_ac_enforcer.AccessPolicyEnforcerLoop
		// captured event handler
		virtualMeshHandler *zephyr_networking_controller.VirtualMeshEventHandlerFuncs
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		virtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshClient = mock_core.NewMockMeshClient(ctrl)
		virtualMeshEventWatcher = mock_zephyr_networking2.NewMockVirtualMeshEventWatcher(ctrl)
		meshEnforcers = []*mock_global_access_control_enforcer.MockAccessPolicyMeshEnforcer{
			mock_global_access_control_enforcer.NewMockAccessPolicyMeshEnforcer(ctrl),
			mock_global_access_control_enforcer.NewMockAccessPolicyMeshEnforcer(ctrl),
		}
		enforcerLoop = global_ac_enforcer.NewEnforcerLoop(
			virtualMeshEventWatcher,
			virtualMeshClient,
			meshClient,
			[]global_ac_enforcer.AccessPolicyMeshEnforcer{
				meshEnforcers[0], meshEnforcers[1],
			},
		)
		virtualMeshEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *zephyr_networking_controller.VirtualMeshEventHandlerFuncs) error {
				virtualMeshHandler = eventHandler
				return nil
			})
		enforcerLoop.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildVirtualMesh = func() *zephyr_networking.VirtualMesh {
		return &zephyr_networking.VirtualMesh{
			Spec: zephyr_networking_types.VirtualMeshSpec{
				Meshes: []*zephyr_core_types.ResourceRef{
					{Name: "name1", Namespace: "namespace1"},
					{Name: "name2", Namespace: "namespace2"},
				},
			},
		}
	}

	var buildMeshes = func() []*zephyr_discovery.Mesh {
		return []*zephyr_discovery.Mesh{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "name1", Namespace: "namespace1"},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "name2", Namespace: "namespace2"},
			},
		}
	}

	It("should start enforcing access control on VirtualMesh creates", func() {
		vm := buildVirtualMesh()
		vm.Spec.EnforceAccessControl = true
		meshes := buildMeshes()
		for i, meshRef := range vm.Spec.GetMeshes() {
			meshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		for _, meshEnforcer := range meshEnforcers {
			meshEnforcer.
				EXPECT().
				StartEnforcing(contextutils.WithLogger(ctx, ""), meshes).
				Return(nil)
			meshEnforcer.
				EXPECT().
				Name().
				Return("")
		}
		var capturedVM *zephyr_networking.VirtualMesh
		virtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *zephyr_networking.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &zephyr_core_types.Status{
			State: zephyr_core_types.Status_ACCEPTED,
		}
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})

	It("should stop enforcing access control on VirtualMesh creates", func() {
		vm := buildVirtualMesh()
		meshes := buildMeshes()
		for i, meshRef := range vm.Spec.GetMeshes() {
			meshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		for _, meshEnforcer := range meshEnforcers {
			meshEnforcer.
				EXPECT().
				StopEnforcing(contextutils.WithLogger(ctx, ""), meshes).
				Return(nil)
			meshEnforcer.
				EXPECT().
				Name().
				Return("")
		}
		var capturedVM *zephyr_networking.VirtualMesh
		virtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *zephyr_networking.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &zephyr_core_types.Status{
			State: zephyr_core_types.Status_ACCEPTED,
		}
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})

	It("should handle errors on VirtualMesh create", func() {
		vm := buildVirtualMesh()
		meshes := buildMeshes()
		testErr := eris.New("err")
		for i, meshRef := range vm.Spec.GetMeshes() {
			meshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		meshEnforcers[0].
			EXPECT().
			StopEnforcing(contextutils.WithLogger(ctx, ""), meshes).
			Return(testErr)
		meshEnforcers[0].
			EXPECT().
			Name().
			Return("")
		var capturedVM *zephyr_networking.VirtualMesh
		virtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *zephyr_networking.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: testErr.Error(),
		}
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})

	It("should clean up everything that's relevant when a virtual mesh is deleted", func() {
		vm := buildVirtualMesh()
		meshes := buildMeshes()
		for i, meshRef := range vm.Spec.GetMeshes() {
			meshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		for _, meshEnforcer := range meshEnforcers {
			meshEnforcer.
				EXPECT().
				StopEnforcing(contextutils.WithLogger(ctx, ""), meshes).
				Return(nil)
			meshEnforcer.
				EXPECT().
				Name().
				Return("")
		}

		err := virtualMeshHandler.DeleteVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
	})
})
