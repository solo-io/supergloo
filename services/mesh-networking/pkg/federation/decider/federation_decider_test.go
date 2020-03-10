package decider_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	mock_discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/decider"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/decider/strategies"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Federation Decider", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("doesn't federate anything for a group with only one member", func() {
		snapshot := snapshot.MeshNetworkingSnapshot{
			MeshGroups: []*networking_v1alpha1.MeshGroup{{
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{{
						Name:      "mesh-1",
						Namespace: env.DefaultWriteNamespace,
					}},
					Federation: &networking_types.Federation{
						Mode: networking_types.Federation_PERMISSIVE,
					},
				},
			}},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)

		meshGroupClient := mock_zephyr_networking.NewMockMeshGroupClient(ctrl)
		groupCopy := *snapshot.MeshGroups[0]
		groupCopy.Status.FederationStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
		meshGroupClient.EXPECT().
			UpdateStatus(ctx, &groupCopy).
			Return(nil)

		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-1",
				Namespace: env.DefaultWriteNamespace,
			}).
			Return(&discovery_v1alpha1.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name:        "mesh-1",
					ClusterName: "cluster-name",
				},
			}, nil)

		decider := decider.NewFederationDecider(meshServiceClient, meshClient, meshGroupClient, func(mode networking_types.Federation_Mode, meshServiceClient discovery_core.MeshServiceClient) (strategies.FederationStrategy, error) {
			return strategies.NewPermissiveFederation(meshServiceClient), nil
		})
		decider.DecideFederation(ctx, &snapshot)
	})

	/************************
	*   This test sets up the following situation:
	*      - we have four meshes, named mesh-n for n from 1 to 4
	*      - mesh-1 through mesh-3 are in a group together, mesh-4 is in a group by itself
	*      - each of the meshes has exactly one mesh service and one mesh workload
	*
	*   We expect that:
	*      - each of the mesh services in group 1 gets federated to the two mesh workloads from the OTHER meshes in group 1
	*      - the service in mesh 4 gets federated nowhere
	*      - nothing gets federated to the workload in mesh 4
	*************************/
	It("federates each service to every other mesh in a group, and not to meshes outside the group (permissive federation end to end)", func() {
		meshService1 := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mesh-service-1-mesh-1",
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-1",
					Namespace: env.DefaultWriteNamespace,
				},
				KubeService: &types.KubeService{
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
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-2",
					Namespace: env.DefaultWriteNamespace,
				},
				KubeService: &types.KubeService{
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
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-3",
					Namespace: env.DefaultWriteNamespace,
				},
				KubeService: &types.KubeService{
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
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name:      "mesh-4",
					Namespace: env.DefaultWriteNamespace,
				},
				KubeService: &types.KubeService{
					Ref: &core_types.ResourceRef{
						Name:      "application-svc4",
						Namespace: "application-ns4",
					},
				},
			},
		}

		snapshot := snapshot.MeshNetworkingSnapshot{
			MeshGroups: []*networking_v1alpha1.MeshGroup{
				{
					Spec: networking_types.MeshGroupSpec{
						Meshes: []*core_types.ResourceRef{
							{
								Name:      "mesh-1",
								Namespace: env.DefaultWriteNamespace,
							},
							{
								Name:      "mesh-2",
								Namespace: env.DefaultWriteNamespace,
							},
							{
								Name:      "mesh-3",
								Namespace: env.DefaultWriteNamespace,
							},
						},
						Federation: nil, // should default to the permissive mode for demo purposes
					},
				},
				{
					Spec: networking_types.MeshGroupSpec{
						Meshes: []*core_types.ResourceRef{{
							Name:      "mesh-4",
							Namespace: env.DefaultWriteNamespace,
						}},
						Federation: &networking_types.Federation{
							Mode: networking_types.Federation_PERMISSIVE,
						},
					},
				},
			},
			MeshServices: []*discovery_v1alpha1.MeshService{meshService1, meshService2, meshService3, meshService4},
			MeshWorkloads: []*discovery_v1alpha1.MeshWorkload{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-1-mesh-1",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-1",
							Namespace: env.DefaultWriteNamespace,
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-2-mesh-2",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-2",
							Namespace: env.DefaultWriteNamespace,
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-3-mesh-3",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-3",
							Namespace: env.DefaultWriteNamespace,
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh-workload-4-mesh-4",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: types.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-4",
							Namespace: env.DefaultWriteNamespace,
						},
					},
				},
			},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)

		meshGroupClient := mock_zephyr_networking.NewMockMeshGroupClient(ctrl)

		// EXPECTs for group 1
		group1Copy := *snapshot.MeshGroups[0]
		group1Copy.Status.FederationStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
		meshGroupClient.EXPECT().
			UpdateStatus(ctx, &group1Copy).
			Return(nil)

		// EXPECTs for group 2
		group2Copy := *snapshot.MeshGroups[1]
		group2Copy.Status.FederationStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
		meshGroupClient.EXPECT().
			UpdateStatus(ctx, &group2Copy).
			Return(nil)

		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshClient.EXPECT().
			Get(ctx, client.ObjectKey{
				Name:      "mesh-1",
				Namespace: env.DefaultWriteNamespace,
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
				Namespace: env.DefaultWriteNamespace,
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
				Namespace: env.DefaultWriteNamespace,
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
				Namespace: env.DefaultWriteNamespace,
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
		meshService1Copy.Spec.Federation = &types.Federation{
			MulticlusterDnsName: "application-svc1.application-ns1.cluster-1",
			FederatedToWorkloads: []*core_types.ResourceRef{
				{
					Name:      "mesh-workload-2-mesh-2",
					Namespace: env.DefaultWriteNamespace,
				},
				{
					Name:      "mesh-workload-3-mesh-3",
					Namespace: env.DefaultWriteNamespace,
				},
			},
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService1Copy).
			Return(nil)

		// EXPECTs for meshService2
		meshService2Copy := *meshService2
		meshService2Copy.Spec.Federation = &types.Federation{
			MulticlusterDnsName: "application-svc2.application-ns2.cluster-2",
			FederatedToWorkloads: []*core_types.ResourceRef{
				{
					Name:      "mesh-workload-1-mesh-1",
					Namespace: env.DefaultWriteNamespace,
				},
				{
					Name:      "mesh-workload-3-mesh-3",
					Namespace: env.DefaultWriteNamespace,
				},
			},
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService2Copy).
			Return(nil)

		// EXPECTs for meshService3
		meshService3Copy := *meshService3
		meshService3Copy.Spec.Federation = &types.Federation{
			MulticlusterDnsName: "application-svc3.application-ns3.cluster-3",
			FederatedToWorkloads: []*core_types.ResourceRef{
				{
					Name:      "mesh-workload-1-mesh-1",
					Namespace: env.DefaultWriteNamespace,
				},
				{
					Name:      "mesh-workload-2-mesh-2",
					Namespace: env.DefaultWriteNamespace,
				},
			},
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService3Copy).
			Return(nil)

		// EXPECTs for meshService4
		meshService4Copy := *meshService4
		meshService4Copy.Spec.Federation = &types.Federation{
			MulticlusterDnsName: "application-svc4.application-ns4.cluster-4",
		}
		meshServiceClient.EXPECT().
			Update(ctx, &meshService4Copy).
			Return(nil)

		decider := decider.NewFederationDecider(meshServiceClient, meshClient, meshGroupClient, strategies.GetFederationStrategyFromMode)
		decider.DecideFederation(ctx, &snapshot)
	})

	It("marks all groups in the snapshot as having a processing error if we can't set up the precomputed data", func() {
		group1 := &networking_v1alpha1.MeshGroup{
			ObjectMeta: v1.ObjectMeta{
				Name: "group-1",
			},
			Spec: networking_types.MeshGroupSpec{
				Meshes: []*core_types.ResourceRef{
					{
						Name:      "mesh-1",
						Namespace: env.DefaultWriteNamespace,
					},
					{
						Name:      "mesh-2",
						Namespace: env.DefaultWriteNamespace,
					},
					{
						Name:      "mesh-3",
						Namespace: env.DefaultWriteNamespace,
					},
				},
				Federation: nil, // should default to the permissive mode for demo purposes
			},
		}
		group2 := &networking_v1alpha1.MeshGroup{
			ObjectMeta: v1.ObjectMeta{
				Name: "group-2",
			},
			Spec: networking_types.MeshGroupSpec{
				Meshes: []*core_types.ResourceRef{{
					Name:      "mesh-4",
					Namespace: env.DefaultWriteNamespace,
				}},
				Federation: &networking_types.Federation{
					Mode: networking_types.Federation_PERMISSIVE,
				},
			},
		}

		snapshot := snapshot.MeshNetworkingSnapshot{
			MeshGroups:    []*networking_v1alpha1.MeshGroup{group1, group2},
			MeshServices:  []*discovery_v1alpha1.MeshService{},
			MeshWorkloads: []*discovery_v1alpha1.MeshWorkload{},
		}

		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshGroupClient := mock_zephyr_networking.NewMockMeshGroupClient(ctrl)

		meshClient.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, testErr).AnyTimes()

		// this many errors happen, because every mesh is going to error out
		var group1MultiErr *multierror.Error
		for range group1.Spec.Meshes {
			group1MultiErr = multierror.Append(group1MultiErr, testErr)
		}
		var group2MultiErr *multierror.Error
		for range group2.Spec.Meshes {
			group2MultiErr = multierror.Append(group2MultiErr, testErr)
		}

		group1Copy := *group1
		group1Copy.Status.FederationStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_PROCESSING_ERROR,
			Message: decider.ErrorLoadingMeshMetadata(group1MultiErr),
		}
		meshGroupClient.EXPECT().
			UpdateStatus(ctx, &group1Copy).
			Return(nil)

		group2Copy := *group2
		group2Copy.Status.FederationStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_PROCESSING_ERROR,
			Message: decider.ErrorLoadingMeshMetadata(group2MultiErr),
		}
		meshGroupClient.EXPECT().
			UpdateStatus(ctx, &group2Copy).
			Return(nil)

		strategyDecider := func(mode networking_types.Federation_Mode, meshServiceClient discovery_core.MeshServiceClient) (strategies.FederationStrategy, error) {
			// these don't matter, we'll bail out before this point
			return nil, nil
		}

		decider.NewFederationDecider(meshServiceClient, meshClient, meshGroupClient, strategyDecider).DecideFederation(ctx, &snapshot)
	})
})
