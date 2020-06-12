package strategies_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	strategies2 "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/strategies"
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
		meshRef := &smh_core_types.ResourceRef{
			Name:      "mesh-1",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		vm := &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{meshRef},
				Federation: &smh_networking_types.VirtualMeshSpec_Federation{
					Mode: smh_networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
				},
			},
		}

		service := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "svc-1",
			},
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: meshRef,
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: &smh_core_types.ResourceRef{
						Name:      "application-svc",
						Namespace: "application-ns",
					},
				},
			},
		}
		perMeshResources := map[string]*strategies2.MeshMetadata{
			"mesh-1": {
				MeshServices: []*smh_discovery.MeshService{service},
				ClusterName:  "application-cluster",
			},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		serviceCopy := *service
		serviceCopy.Spec.Federation = &smh_discovery_types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc.application-ns.application-cluster",
		}
		meshServiceClient.EXPECT().
			UpsertMeshService(ctx, &serviceCopy).
			Return(nil)

		err := strategies2.NewPermissiveFederation(meshServiceClient).WriteFederationToServices(ctx, vm, perMeshResources)
		Expect(err).NotTo(HaveOccurred())
	})
})
