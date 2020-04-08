package decider_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/decider"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/decider/strategies"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/snapshot"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		snapshot := snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*networking_v1alpha1.VirtualMesh{{
				Spec: networking_types.VirtualMeshSpec{
					Meshes: []*core_types.ResourceRef{{
						Name:      "mesh-1",
						Namespace: env.GetWriteNamespace(),
					}},
					Federation: &networking_types.VirtualMeshSpec_Federation{
						Mode: networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
					},
				},
			}},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)

		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		vmCopy := *snapshot.VirtualMeshes[0]
		vmCopy.Status.FederationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, &vmCopy).
			Return(nil)

		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-1",
				Namespace: env.GetWriteNamespace(),
			}).
			Return(&discovery_v1alpha1.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name:        "mesh-1",
					ClusterName: "cluster-name",
				},
			}, nil)

		decider := decider.NewFederationDecider(meshServiceClient, meshClient, virtualMeshClient, func(mode networking_types.VirtualMeshSpec_Federation_Mode, meshServiceClient discovery_core.MeshServiceClient) (strategies.FederationStrategy, error) {
			return strategies.NewPermissiveFederation(meshServiceClient), nil
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
		meshService1 := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mesh-service-1-mesh-1",
				Namespace: env.GetWriteNamespace(),
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &core_types.ResourceRef{
						Name:      "application-svc1",
						Namespace: "application-ns1",
					},
				},
			},
		}
		meshService2 := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mesh-service-2-mesh-2",
				Namespace: env.GetWriteNamespace(),
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-2",
					Namespace: env.GetWriteNamespace(),
				},
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &core_types.ResourceRef{
						Name:      "application-svc2",
						Namespace: "application-ns2",
					},
				},
			},
		}
		meshService3 := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mesh-service-3-mesh-3",
				Namespace: env.GetWriteNamespace(),
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-3",
					Namespace: env.GetWriteNamespace(),
				},
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &core_types.ResourceRef{
						Name:      "application-svc3",
						Namespace: "application-ns3",
					},
				},
			},
		}
		meshService4 := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mesh-service-4-mesh-4",
				Namespace: env.GetWriteNamespace(),
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-4",
					Namespace: env.GetWriteNamespace(),
				},
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &core_types.ResourceRef{
						Name:      "application-svc4",
						Namespace: "application-ns4",
					},
				},
			},
		}

		snapshot := snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*networking_v1alpha1.VirtualMesh{
				{
					Spec: networking_types.VirtualMeshSpec{
						Meshes: []*core_types.ResourceRef{
							{
								Name:      "mesh-1",
								Namespace: env.GetWriteNamespace(),
							},
							{
								Name:      "mesh-2",
								Namespace: env.GetWriteNamespace(),
							},
							{
								Name:      "mesh-3",
								Namespace: env.GetWriteNamespace(),
							},
						},
						Federation: nil, // should default to the permissive mode for demo purposes
					},
				},
				{
					Spec: networking_types.VirtualMeshSpec{
						Meshes: []*core_types.ResourceRef{{
							Name:      "mesh-4",
							Namespace: env.GetWriteNamespace(),
						}},
						Federation: &networking_types.VirtualMeshSpec_Federation{
							Mode: networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
						},
					},
				},
			},
			MeshServices: []*discovery_v1alpha1.MeshService{meshService1, meshService2, meshService3, meshService4},
			MeshWorkloads: []*discovery_v1alpha1.MeshWorkload{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-1-mesh-1",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-1",
							Namespace: env.GetWriteNamespace(),
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-2-mesh-2",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-2",
							Namespace: env.GetWriteNamespace(),
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-3-mesh-3",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-3",
							Namespace: env.GetWriteNamespace(),
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-4-mesh-4",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-4",
							Namespace: env.GetWriteNamespace(),
						},
					},
				},
			},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)

		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)

		// EXPECTs for vm 1
		vm1Copy := *snapshot.VirtualMeshes[0]
		vm1Copy.Status.FederationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, &vm1Copy).
			Return(nil)

		// EXPECTs for vm 2
		vm2Copy := *snapshot.VirtualMeshes[1]
		vm2Copy.Status.FederationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, &vm2Copy).
			Return(nil)

		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-1",
				Namespace: env.GetWriteNamespace(),
			}).
			Return(&discovery_v1alpha1.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-1",
				},
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "cluster-1",
					},
				},
			}, nil)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-2",
				Namespace: env.GetWriteNamespace(),
			}).
			Return(&discovery_v1alpha1.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-2",
				},
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "cluster-2",
					},
				},
			}, nil)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-3",
				Namespace: env.GetWriteNamespace(),
			}).
			Return(&discovery_v1alpha1.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-3",
				},
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "cluster-3",
					},
				},
			}, nil)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-4",
				Namespace: env.GetWriteNamespace(),
			}).
			Return(&discovery_v1alpha1.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-4",
				},
				Spec: types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "cluster-4",
					},
				},
			}, nil)

		// EXPECTs for meshService1
		meshService1Copy := *meshService1
		meshService1Copy.Spec.Federation = &types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc1.application-ns1.cluster-1",
			FederatedToWorkloads: []*core_types.ResourceRef{
				{
					Name:      "mesh-workload-2-mesh-2",
					Namespace: env.GetWriteNamespace(),
				},
				{
					Name:      "mesh-workload-3-mesh-3",
					Namespace: env.GetWriteNamespace(),
				},
			},
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService1Copy).
			Return(nil)

		// EXPECTs for meshService2
		meshService2Copy := *meshService2
		meshService2Copy.Spec.Federation = &types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc2.application-ns2.cluster-2",
			FederatedToWorkloads: []*core_types.ResourceRef{
				{
					Name:      "mesh-workload-1-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				{
					Name:      "mesh-workload-3-mesh-3",
					Namespace: env.GetWriteNamespace(),
				},
			},
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService2Copy).
			Return(nil)

		// EXPECTs for meshService3
		meshService3Copy := *meshService3
		meshService3Copy.Spec.Federation = &types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc3.application-ns3.cluster-3",
			FederatedToWorkloads: []*core_types.ResourceRef{
				{
					Name:      "mesh-workload-1-mesh-1",
					Namespace: env.GetWriteNamespace(),
				},
				{
					Name:      "mesh-workload-2-mesh-2",
					Namespace: env.GetWriteNamespace(),
				},
			},
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService3Copy).
			Return(nil)

		// EXPECTs for meshService4
		meshService4Copy := *meshService4
		meshService4Copy.Spec.Federation = &types.MeshServiceSpec_Federation{
			MulticlusterDnsName: "application-svc4.application-ns4.cluster-4",
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService4Copy).
			Return(nil)

		decider := decider.NewFederationDecider(meshServiceClient, meshClient, virtualMeshClient, strategies.GetFederationStrategyFromMode)
		decider.DecideFederation(ctx, &snapshot)
	})

	It("marks all virtual meshes in the snapshot as having a processing error if we can't set up the precomputed data", func() {
		vm1 := &networking_v1alpha1.VirtualMesh{
			ObjectMeta: v1.ObjectMeta{
				Name: "virtual-mesh-1",
			},
			Spec: networking_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{
					{
						Name:      "mesh-1",
						Namespace: env.GetWriteNamespace(),
					},
					{
						Name:      "mesh-2",
						Namespace: env.GetWriteNamespace(),
					},
					{
						Name:      "mesh-3",
						Namespace: env.GetWriteNamespace(),
					},
				},
				Federation: nil, // should default to the permissive mode for demo purposes
			},
		}
		vm2 := &networking_v1alpha1.VirtualMesh{
			ObjectMeta: v1.ObjectMeta{
				Name: "virtual-mesh-2",
			},
			Spec: networking_types.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{{
					Name:      "mesh-4",
					Namespace: env.GetWriteNamespace(),
				}},
				Federation: &networking_types.VirtualMeshSpec_Federation{
					Mode: networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
				},
			},
		}

		snapshot := snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*networking_v1alpha1.VirtualMesh{vm1, vm2},
			MeshServices:  []*discovery_v1alpha1.MeshService{},
			MeshWorkloads: []*discovery_v1alpha1.MeshWorkload{},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)

		meshClient.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, testErr).AnyTimes()

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
		vm1Copy.Status.FederationStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: decider.ErrorLoadingMeshMetadata(vm1MultiErr),
		}
		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, &vm1Copy).
			Return(nil)

		vm2Copy := *vm2
		vm2Copy.Status.FederationStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: decider.ErrorLoadingMeshMetadata(vm2MultiErr),
		}
		virtualMeshClient.EXPECT().
			UpdateStatus(ctx, &vm2Copy).
			Return(nil)

		strategyDecider := func(mode networking_types.VirtualMeshSpec_Federation_Mode, meshServiceClient discovery_core.MeshServiceClient) (strategies.FederationStrategy, error) {
			// these don't matter, we'll bail out before this point
			return nil, nil
		}

		decider.NewFederationDecider(meshServiceClient, meshClient, virtualMeshClient, strategyDecider).DecideFederation(ctx, &snapshot)
	})
})
