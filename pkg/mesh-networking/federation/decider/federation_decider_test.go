package decider_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	strategies2 "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/strategies"
	networking_snapshot "github.com/solo-io/service-mesh-hub/pkg/common/snapshot"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/decider"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Federation Decider", func() {
	var (
		ctrl   *gomock.Controller
		ctx    context.Context
		logger *test_logging.TestLogger

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		logger = test_logging.NewTestLogger()
		ctrl = gomock.NewController(GinkgoT())
		ctx = contextutils.WithExistingLogger(ctx, logger.Logger())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("doesn't federate anything for a virtual mesh with only one member", func() {
		snapshot := networking_snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*smh_networking.VirtualMesh{{
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{{
						Name:      "mesh-1",
						Namespace: container_runtime.GetWriteNamespace(),
					}},
					Federation: &smh_networking_types.VirtualMeshSpec_Federation{
						Mode: smh_networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
					},
				},
			}},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)

		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		vmCopy := *snapshot.VirtualMeshes[0]
		vmCopy.Status.FederationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, &vmCopy).
			Return(nil)

		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshClient.EXPECT().
			GetMesh(ctx, client.ObjectKey{
				Name:      "mesh-1",
				Namespace: container_runtime.GetWriteNamespace(),
			}).
			Return(&smh_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:        "mesh-1",
					ClusterName: "cluster-name",
				},
			}, nil)

		decider := decider.NewFederationDecider(meshServiceClient, meshClient, virtualMeshClient, func(mode smh_networking_types.VirtualMeshSpec_Federation_Mode, meshServiceClient smh_discovery.MeshServiceClient) (strategies2.FederationStrategy, error) {
			return strategies2.NewPermissiveFederation(meshServiceClient), nil
		})
		decider.DecideFederation(ctx, &snapshot)
	})

	/************************
	*   This test sets up the following situation:
	*      - we have four meshes, named mesh-n for n from 1 to 4
	*      - mesh-1 through mesh-3 are in a vm together, mesh-4 is in a vm by itself
	*      - each of the meshes has exactly one mesh service and one mesh workload
	*
	*   We expect that:
	*      - each of the mesh services in vm 1 gets federated to the two mesh workloads from the OTHER meshes in vm 1
	*      - the service in mesh 4 gets federated nowhere
	*      - nothing gets federated to the workload in mesh 4
	*************************/
	It("federates each service to every other mesh in a vm, and not to meshes outside the vm (permissive federation end to end)", func() {
		meshService1 := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "mesh-service-1-mesh-1",
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name:      "mesh-1",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: &smh_core_types.ResourceRef{
						Name:      "application-svc1",
						Namespace: "application-ns1",
					},
				},
			},
		}
		meshService2 := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "mesh-service-2-mesh-2",
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name:      "mesh-2",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: &smh_core_types.ResourceRef{
						Name:      "application-svc2",
						Namespace: "application-ns2",
					},
				},
			},
		}
		meshService3 := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "mesh-service-3-mesh-3",
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name:      "mesh-3",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: &smh_core_types.ResourceRef{
						Name:      "application-svc3",
						Namespace: "application-ns3",
					},
				},
			},
		}
		meshService4 := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "mesh-service-4-mesh-4",
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name:      "mesh-4",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: &smh_core_types.ResourceRef{
						Name:      "application-svc4",
						Namespace: "application-ns4",
					},
				},
			},
		}

		snapshot := networking_snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*smh_networking.VirtualMesh{
				{
					Spec: smh_networking_types.VirtualMeshSpec{
						Meshes: []*smh_core_types.ResourceRef{
							{
								Name:      "mesh-1",
								Namespace: container_runtime.GetWriteNamespace(),
							},
							{
								Name:      "mesh-2",
								Namespace: container_runtime.GetWriteNamespace(),
							},
							{
								Name:      "mesh-3",
								Namespace: container_runtime.GetWriteNamespace(),
							},
						},
						Federation: nil, // should default to the permissive mode for demo purposes
					},
				},
				{
					Spec: smh_networking_types.VirtualMeshSpec{
						Meshes: []*smh_core_types.ResourceRef{{
							Name:      "mesh-4",
							Namespace: container_runtime.GetWriteNamespace(),
						}},
						Federation: &smh_networking_types.VirtualMeshSpec_Federation{
							Mode: smh_networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
						},
					},
				},
			},
			MeshServices: []*smh_discovery.MeshService{meshService1, meshService2, meshService3, meshService4},
			MeshWorkloads: []*smh_discovery.MeshWorkload{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-workload-1-mesh-1",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshWorkloadSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-1",
							Namespace: container_runtime.GetWriteNamespace(),
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-workload-2-mesh-2",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshWorkloadSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-2",
							Namespace: container_runtime.GetWriteNamespace(),
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-workload-3-mesh-3",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshWorkloadSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-3",
							Namespace: container_runtime.GetWriteNamespace(),
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-workload-4-mesh-4",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshWorkloadSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-4",
							Namespace: container_runtime.GetWriteNamespace(),
						},
					},
				},
			},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)

		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)

		// EXPECTs for vm 1
		vm1Copy := *snapshot.VirtualMeshes[0]
		vm1Copy.Status.FederationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, &vm1Copy).
			Return(nil)

		// EXPECTs for vm 2
		vm2Copy := *snapshot.VirtualMeshes[1]
		vm2Copy.Status.FederationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, &vm2Copy).
			Return(nil)

		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshClient.EXPECT().
			GetMesh(ctx, client.ObjectKey{
				Name:      "mesh-1",
				Namespace: container_runtime.GetWriteNamespace(),
			}).
			Return(&smh_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "mesh-1",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "cluster-1",
					},
				},
			}, nil)
		meshClient.EXPECT().
			GetMesh(ctx, client.ObjectKey{
				Name:      "mesh-2",
				Namespace: container_runtime.GetWriteNamespace(),
			}).
			Return(&smh_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "mesh-2",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "cluster-2",
					},
				},
			}, nil)
		meshClient.EXPECT().
			GetMesh(ctx, client.ObjectKey{
				Name:      "mesh-3",
				Namespace: container_runtime.GetWriteNamespace(),
			}).
			Return(&smh_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "mesh-3",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "cluster-3",
					},
				},
			}, nil)
		meshClient.EXPECT().
			GetMesh(ctx, client.ObjectKey{
				Name:      "mesh-4",
				Namespace: container_runtime.GetWriteNamespace(),
			}).
			Return(&smh_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "mesh-4",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "cluster-4",
					},
				},
			}, nil)

		// EXPECTs for meshService1
		meshService1Copy := *meshService1
		meshService1Copy.Spec.Federation = &smh_discovery_types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc1.application-ns1.cluster-1",
			FederatedToWorkloads: []*smh_core_types.ResourceRef{
				{
					Name:      "mesh-workload-2-mesh-2",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				{
					Name:      "mesh-workload-3-mesh-3",
					Namespace: container_runtime.GetWriteNamespace(),
				},
			},
		}
		meshServiceClient.EXPECT().
			UpsertMeshServiceSpec(ctx, &meshService1Copy).
			Return(nil)

		// EXPECTs for meshService2
		meshService2Copy := *meshService2
		meshService2Copy.Spec.Federation = &smh_discovery_types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc2.application-ns2.cluster-2",
			FederatedToWorkloads: []*smh_core_types.ResourceRef{
				{
					Name:      "mesh-workload-1-mesh-1",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				{
					Name:      "mesh-workload-3-mesh-3",
					Namespace: container_runtime.GetWriteNamespace(),
				},
			},
		}
		meshServiceClient.EXPECT().
			UpsertMeshServiceSpec(ctx, &meshService2Copy).
			Return(nil)

		// EXPECTs for meshService3
		meshService3Copy := *meshService3
		meshService3Copy.Spec.Federation = &smh_discovery_types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc3.application-ns3.cluster-3",
			FederatedToWorkloads: []*smh_core_types.ResourceRef{
				{
					Name:      "mesh-workload-1-mesh-1",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				{
					Name:      "mesh-workload-2-mesh-2",
					Namespace: container_runtime.GetWriteNamespace(),
				},
			},
		}
		meshServiceClient.EXPECT().
			UpsertMeshServiceSpec(ctx, &meshService3Copy).
			Return(nil)

		// EXPECTs for meshService4
		meshService4Copy := *meshService4
		meshService4Copy.Spec.Federation = &smh_discovery_types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc4.application-ns4.cluster-4",
		}
		meshServiceClient.EXPECT().
			UpsertMeshServiceSpec(ctx, &meshService4Copy).
			Return(nil)

		decider := decider.NewFederationDecider(meshServiceClient, meshClient, virtualMeshClient, strategies2.GetFederationStrategyFromMode)
		decider.DecideFederation(ctx, &snapshot)
	})

	It("marks all virtual meshes in the snapshot as having a processing error if we can't set up the precomputed data", func() {
		vm1 := &smh_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "virtual-mesh-1",
			},
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{
					{
						Name:      "mesh-1",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					{
						Name:      "mesh-2",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					{
						Name:      "mesh-3",
						Namespace: container_runtime.GetWriteNamespace(),
					},
				},
				Federation: nil, // should default to the permissive mode for demo purposes
			},
		}
		vm2 := &smh_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "virtual-mesh-2",
			},
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{{
					Name:      "mesh-4",
					Namespace: container_runtime.GetWriteNamespace(),
				}},
				Federation: &smh_networking_types.VirtualMeshSpec_Federation{
					Mode: smh_networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
				},
			},
		}

		snapshot := networking_snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*smh_networking.VirtualMesh{vm1, vm2},
			MeshServices:  []*smh_discovery.MeshService{},
			MeshWorkloads: []*smh_discovery.MeshWorkload{},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)

		meshClient.EXPECT().GetMesh(gomock.Any(), gomock.Any()).Return(nil, testErr).AnyTimes()

		// this many errors happen, because every mesh is going to error out
		var vm1MultiErr *multierror.Error
		for range vm1.Spec.Meshes {
			vm1MultiErr = multierror.Append(vm1MultiErr, testErr)
		}
		var vm2MultiErr *multierror.Error
		for range vm2.Spec.Meshes {
			vm2MultiErr = multierror.Append(vm2MultiErr, testErr)
		}

		vm1Copy := *vm1
		vm1Copy.Status.FederationStatus = &smh_core_types.Status{
			State:   smh_core_types.Status_PROCESSING_ERROR,
			Message: decider.ErrorLoadingMeshMetadata(vm1MultiErr),
		}
		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, &vm1Copy).
			Return(nil)

		vm2Copy := *vm2
		vm2Copy.Status.FederationStatus = &smh_core_types.Status{
			State:   smh_core_types.Status_PROCESSING_ERROR,
			Message: decider.ErrorLoadingMeshMetadata(vm2MultiErr),
		}
		virtualMeshClient.EXPECT().
			UpdateVirtualMeshStatus(ctx, &vm2Copy).
			Return(nil)

		strategyDecider := func(mode smh_networking_types.VirtualMeshSpec_Federation_Mode, meshServiceClient smh_discovery.MeshServiceClient) (strategies2.FederationStrategy, error) {
			// these don't matter, we'll bail out before this point
			return nil, nil
		}

		decider.NewFederationDecider(meshServiceClient, meshClient, virtualMeshClient, strategyDecider).DecideFederation(ctx, &snapshot)
	})
})
