package access_policy_enforcer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	global_ac_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer"
	mock_global_access_control_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking2 "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/networking"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("EnforcerLoop", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		virtualMeshController *mock_zephyr_networking2.MockVirtualMeshEventWatcher
		virtualMeshClient     *mock_zephyr_networking.MockVirtualMeshClient
		meshClient            *mock_core.MockMeshClient
		meshEnforcers         []*mock_global_access_control_enforcer.MockAccessPolicyMeshEnforcer
		enforcerLoop          global_ac_enforcer.AccessPolicyEnforcerLoop
		// captured event handler
		virtualMeshHandler *networking_controller.VirtualMeshEventHandlerFuncs
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		virtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshClient = mock_core.NewMockMeshClient(ctrl)
		virtualMeshController = mock_zephyr_networking2.NewMockVirtualMeshEventWatcher(ctrl)
		meshEnforcers = []*mock_global_access_control_enforcer.MockAccessPolicyMeshEnforcer{
			mock_global_access_control_enforcer.NewMockAccessPolicyMeshEnforcer(ctrl),
			mock_global_access_control_enforcer.NewMockAccessPolicyMeshEnforcer(ctrl),
		}
		enforcerLoop = global_ac_enforcer.NewEnforcerLoop(
			virtualMeshController,
			virtualMeshClient,
			meshClient,
			[]global_ac_enforcer.AccessPolicyMeshEnforcer{
				meshEnforcers[0], meshEnforcers[1],
			},
		)
		virtualMeshController.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *networking_controller.VirtualMeshEventHandlerFuncs) error {
				virtualMeshHandler = eventHandler
				return nil
			})
		enforcerLoop.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildVirtualMesh = func() *networking_v1alpha1.VirtualMesh {
		return &networking_v1alpha1.VirtualMesh{
			Spec: networking_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{
					{Name: "name1", Namespace: "namespace1"},
					{Name: "name2", Namespace: "namespace2"},
				},
			},
		}
	}

	var buildMeshes = func() []*discovery_v1alpha1.Mesh {
		return []*discovery_v1alpha1.Mesh{
			{
				ObjectMeta: v1.ObjectMeta{Name: "name1", Namespace: "namespace1"},
			},
			{
				ObjectMeta: v1.ObjectMeta{Name: "name2", Namespace: "namespace2"},
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
				GetMesh(ctx, clients.ResourceRefToObjectKey(meshRef)).
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
		var capturedVM *networking_v1alpha1.VirtualMesh
		virtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *networking_v1alpha1.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &core_types.Status{
			State: core_types.Status_ACCEPTED,
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
				GetMesh(ctx, clients.ResourceRefToObjectKey(meshRef)).
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
		var capturedVM *networking_v1alpha1.VirtualMesh
		virtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *networking_v1alpha1.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &core_types.Status{
			State: core_types.Status_ACCEPTED,
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
				GetMesh(ctx, clients.ResourceRefToObjectKey(meshRef)).
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
		var capturedVM *networking_v1alpha1.VirtualMesh
		virtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *networking_v1alpha1.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: testErr.Error(),
		}
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})
})
