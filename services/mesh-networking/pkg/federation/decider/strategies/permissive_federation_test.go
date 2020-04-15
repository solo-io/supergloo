package strategies_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/decider/strategies"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Permissive Federation", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("doesn't federate anything for a virtual mesh with only one member", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "mesh-1",
			Namespace: env.GetWriteNamespace(),
		}
		vm := &zephyr_networking.VirtualMesh{
			Spec: networking_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{meshRef},
				Federation: &networking_types.VirtualMeshSpec_Federation{
					Mode: networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
				},
			},
		}

		service := &zephyr_discovery.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name: "svc-1",
			},
			Spec: types.MeshServiceSpec{
				Mesh: meshRef,
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &core_types.ResourceRef{
						Name:      "application-svc",
						Namespace: "application-ns",
					},
				},
			},
		}
		perMeshResources := map[string]*strategies.MeshMetadata{
			"mesh-1": {
				MeshServices: []*zephyr_discovery.MeshService{service},
				ClusterName:  "application-cluster",
			},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		serviceCopy := *service
		serviceCopy.Spec.Federation = &types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc.application-ns.application-cluster",
		}
		meshServiceClient.EXPECT().
			UpdateMeshService(ctx, &serviceCopy).
			Return(nil)

		err := strategies.NewPermissiveFederation(meshServiceClient).WriteFederationToServices(ctx, vm, perMeshResources)
		Expect(err).NotTo(HaveOccurred())
	})
})
