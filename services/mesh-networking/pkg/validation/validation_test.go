package group_validation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	v1alpha1_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	group_validation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/validation"
	mock_group_validation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/validation/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validator", func() {

	var (
		ctrl            *gomock.Controller
		validator       snapshot.MeshNetworkingSnapshotValidator
		meshFinder      *mock_group_validation.MockGroupMeshFinder
		meshGroupClient *mock_zephyr_networking.MockMeshGroupClient
		ctx             context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		meshFinder = mock_group_validation.NewMockGroupMeshFinder(ctrl)
		meshGroupClient = mock_zephyr_networking.NewMockMeshGroupClient(ctrl)
		validator = group_validation.NewMeshGroupValidator(meshFinder, meshGroupClient)
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
		mg := &networkingv1alpha1.MeshGroup{
			Spec: v1alpha1_types.MeshGroupSpec{
				Meshes: []*core_types.ResourceRef{ref},
			},
			Status: v1alpha1_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: testErr.Error(),
				},
			},
		}
		meshFinder.EXPECT().
			GetMeshesForGroup(ctx, mg).
			Return(nil, testErr)

		meshGroupClient.EXPECT().
			UpdateStatus(ctx, mg).
			Return(nil)

		valid := validator.ValidateMeshGroupUpsert(ctx, mg, nil)
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
		mg := &networkingv1alpha1.MeshGroup{
			Spec: v1alpha1_types.MeshGroupSpec{
				Meshes: []*core_types.ResourceRef{ref},
			},
			Status: v1alpha1_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: group_validation.OnlyIstioSupportedError(mesh.Name).Error(),
				},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForGroup(ctx, mg).
			Return([]*discoveryv1alpha1.Mesh{&mesh}, nil)

		meshGroupClient.EXPECT().
			UpdateStatus(ctx, mg).
			Return(nil)

		valid := validator.ValidateMeshGroupUpsert(ctx, mg, nil)
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
					Istio: &discovery_types.IstioMesh{},
				},
			},
		}
		mg := &networkingv1alpha1.MeshGroup{
			Spec: v1alpha1_types.MeshGroupSpec{
				Meshes: []*core_types.ResourceRef{ref},
			},
		}

		meshFinder.EXPECT().
			GetMeshesForGroup(ctx, mg).
			Return([]*discoveryv1alpha1.Mesh{&mesh}, nil)

		valid := validator.ValidateMeshGroupUpsert(ctx, mg, nil)
		Expect(valid).To(BeTrue())
	})
})
