package strategies_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/decider/strategies"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "mesh-1",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		vm := &zephyr_networking.VirtualMesh{
			Spec: zephyr_networking_types.VirtualMeshSpec{
				Meshes: []*zephyr_core_types.ResourceRef{meshRef},
				Federation: &zephyr_networking_types.VirtualMeshSpec_Federation{
					Mode: zephyr_networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
				},
			},
		}

		service := &zephyr_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "svc-1",
			},
			Spec: zephyr_discovery_types.MeshServiceSpec{
				Mesh: meshRef,
				KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
					Ref: &zephyr_core_types.ResourceRef{
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
		serviceCopy.Spec.Federation = &zephyr_discovery_types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc.application-ns.application-cluster",
		}
		meshServiceClient.EXPECT().
			UpsertMeshServiceSpec(ctx, &serviceCopy).
			Return(nil)

		err := strategies.NewPermissiveFederation(meshServiceClient).WriteFederationToServices(ctx, vm, perMeshResources)
		Expect(err).NotTo(HaveOccurred())
	})
})
