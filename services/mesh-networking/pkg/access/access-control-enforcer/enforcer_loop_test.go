package access_policy_enforcer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/testutils"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/common/access-control/enforcer"
	mock_access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/common/access-control/enforcer/mocks"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	global_ac_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	mock_smh_networking2 "github.com/solo-io/service-mesh-hub/test/mocks/smh/networking"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("EnforcerLoop", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockVirtualMeshEventWatcher *mock_smh_networking2.MockVirtualMeshEventWatcher
		mockVirtualMeshClient       *mock_smh_networking.MockVirtualMeshClient
		mockMeshClient              *mock_core.MockMeshClient
		mockMeshEnforcers           []*mock_access_control_enforcer.MockAccessPolicyMeshEnforcer
		enforcerLoop                global_ac_enforcer.AccessPolicyEnforcerLoop
		// captured event handler
		virtualMeshHandler *smh_networking_controller.VirtualMeshEventHandlerFuncs
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockVirtualMeshClient = mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockVirtualMeshEventWatcher = mock_smh_networking2.NewMockVirtualMeshEventWatcher(ctrl)
		mockMeshEnforcers = []*mock_access_control_enforcer.MockAccessPolicyMeshEnforcer{
			mock_access_control_enforcer.NewMockAccessPolicyMeshEnforcer(ctrl),
			mock_access_control_enforcer.NewMockAccessPolicyMeshEnforcer(ctrl),
		}
		enforcerLoop = global_ac_enforcer.NewEnforcerLoop(
			mockVirtualMeshEventWatcher,
			mockVirtualMeshClient,
			mockMeshClient,
			[]access_control_enforcer.AccessPolicyMeshEnforcer{
				mockMeshEnforcers[0], mockMeshEnforcers[1],
			},
		)
		mockVirtualMeshEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_networking_controller.VirtualMeshEventHandlerFuncs) error {
				virtualMeshHandler = eventHandler
				return nil
			})
		enforcerLoop.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildVirtualMesh = func() *smh_networking.VirtualMesh {
		return &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{
					{Name: "name1", Namespace: "namespace1"},
					{Name: "name2", Namespace: "namespace2"},
				},
			},
		}
	}

	var buildMeshesWithDefaultAccessControlEnabled = func() []*smh_discovery.Mesh {
		return []*smh_discovery.Mesh{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "name1", Namespace: "namespace1"},
				Spec: types.MeshSpec{
					MeshType: &types.MeshSpec_Istio1_5_{},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "name2", Namespace: "namespace2"},
				Spec: types.MeshSpec{
					MeshType: &types.MeshSpec_AwsAppMesh_{},
				},
			},
		}
	}

	// Temporary, see function definition for "cleanupAppmeshResources" in enforcer_loop.go for more details
	var expectAppmeshCleanup = func() {
		mockMeshClient.EXPECT().ListMesh(ctx).Return(&smh_discovery.MeshList{}, nil)
		mockVirtualMeshClient.EXPECT().ListVirtualMesh(ctx).Return(&smh_networking.VirtualMeshList{}, nil)
	}

	It("should start enforcing access control on VirtualMesh creates", func() {
		vm := buildVirtualMesh()
		vm.Spec.EnforceAccessControl = smh_networking_types.VirtualMeshSpec_ENABLED
		mockVirtualMeshClient.
			EXPECT().
			ListVirtualMesh(ctx).
			Return(&smh_networking.VirtualMeshList{Items: []smh_networking.VirtualMesh{*vm}}, nil)
		meshes := buildMeshesWithDefaultAccessControlEnabled()
		for i, meshRef := range vm.Spec.GetMeshes() {
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		for _, mesh := range meshes {
			for _, meshEnforcer := range mockMeshEnforcers {
				meshEnforcer.
					EXPECT().
					ReconcileAccessControl(contextutils.WithLogger(ctx, ""), mesh, vm).
					Return(nil)
				meshEnforcer.
					EXPECT().
					Name().
					Return("")
			}
		}

		var capturedVM *smh_networking.VirtualMesh
		mockVirtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *smh_networking.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		expectAppmeshCleanup()
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})

	It("should stop enforcing access control on VirtualMesh creates", func() {
		vm := buildVirtualMesh()
		vm.Spec.EnforceAccessControl = smh_networking_types.VirtualMeshSpec_DISABLED
		mockVirtualMeshClient.
			EXPECT().
			ListVirtualMesh(ctx).
			Return(&smh_networking.VirtualMeshList{Items: []smh_networking.VirtualMesh{*vm}}, nil)
		meshes := buildMeshesWithDefaultAccessControlEnabled()
		for i, meshRef := range vm.Spec.GetMeshes() {
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		for _, mesh := range meshes {
			for _, meshEnforcer := range mockMeshEnforcers {
				meshEnforcer.
					EXPECT().
					ReconcileAccessControl(contextutils.WithLogger(ctx, ""), mesh, vm).
					Return(nil)
				meshEnforcer.
					EXPECT().
					Name().
					Return("")
			}
		}
		var capturedVM *smh_networking.VirtualMesh
		mockVirtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *smh_networking.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		expectAppmeshCleanup()
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})

	It("should handle errors on VirtualMesh create", func() {
		vm := buildVirtualMesh()
		mockVirtualMeshClient.
			EXPECT().
			ListVirtualMesh(ctx).
			Return(&smh_networking.VirtualMeshList{Items: []smh_networking.VirtualMesh{*vm}}, nil)

		meshes := buildMeshesWithDefaultAccessControlEnabled()
		testErr := eris.New("err")
		for i, meshRef := range vm.Spec.GetMeshes() {
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}
		mockMeshEnforcers[0].
			EXPECT().
			ReconcileAccessControl(contextutils.WithLogger(ctx, ""), meshes[0], vm).
			Return(testErr)
		mockMeshEnforcers[0].
			EXPECT().
			Name().
			Return("")
		var capturedVM *smh_networking.VirtualMesh
		mockVirtualMeshClient.
			EXPECT().
			UpdateVirtualMeshStatus(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, virtualMesh *smh_networking.VirtualMesh) error {
				capturedVM = virtualMesh
				return nil
			})
		expectedVMStatus := &smh_core_types.Status{
			State:   smh_core_types.Status_PROCESSING_ERROR,
			Message: testErr.Error(),
		}
		expectAppmeshCleanup()
		err := virtualMeshHandler.CreateVirtualMesh(vm)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(capturedVM.Status.AccessControlEnforcementStatus).To(Equal(expectedVMStatus))
	})

	It("should clean up translated resources when a virtual mesh is deleted and stop enforcing if default access control is false", func() {
		vm := buildVirtualMesh()
		mockVirtualMeshClient.
			EXPECT().
			ListVirtualMesh(ctx).
			Return(&smh_networking.VirtualMeshList{Items: []smh_networking.VirtualMesh{*vm}}, nil)
		meshes := buildMeshesWithDefaultAccessControlEnabled()
		for i, meshRef := range vm.Spec.GetMeshes() {
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef)).
				Return(meshes[i], nil)
		}

		// Istio should stop enforcing by default
		for _, meshEnforcer := range mockMeshEnforcers {
			meshEnforcer.
				EXPECT().
				ReconcileAccessControl(contextutils.WithLogger(ctx, ""), meshes[0], vm).
				Return(nil)
			meshEnforcer.
				EXPECT().
				Name().
				Return("")
		}
		// Appmesh should start enforcing by default
		for _, meshEnforcer := range mockMeshEnforcers {
			meshEnforcer.
				EXPECT().
				ReconcileAccessControl(contextutils.WithLogger(ctx, ""), meshes[1], vm).
				Return(nil)
			meshEnforcer.
				EXPECT().
				Name().
				Return("")
		}
		expectAppmeshCleanup()
		err := virtualMeshHandler.DeleteVirtualMesh(vm)
		Expect(err).ToNot(HaveOccurred())
	})
})
