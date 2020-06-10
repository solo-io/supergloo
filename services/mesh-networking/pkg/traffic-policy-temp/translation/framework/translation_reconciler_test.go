package translation_framework_test

import (
	"context"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	translation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	mock_snapshot "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot/mocks"
	mock_zephyr_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	istio_networking "istio.io/api/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TranslationReconciler", func() {
	var (
		ctx  context.Context
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("no resources exist", func() {
		It("does nothing", func() {
			meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
			meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
			reconciler := translation_framework.NewTranslationReconciler(meshServiceClient, meshClient, nil, nil)

			meshClient.EXPECT().
				ListMesh(ctx).
				Return(&zephyr_discovery.MeshList{}, nil)

			err := reconciler.Reconcile(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("we have meshes", func() {
		When("we have no traffic targets (services or workloads) on those meshes", func() {
			It("still runs the output reconciliation", func() {
				meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
				meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
				snapshotReconciler := mock_snapshot.NewMockTranslationSnapshotReconciler(ctrl)
				reconciler := translation_framework.NewTranslationReconciler(meshServiceClient, meshClient, nil, snapshotReconciler)
				knownMeshes := []*zephyr_discovery.Mesh{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-1"},
						Spec: zephyr_discovery_types.MeshSpec{
							MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-2"},
						Spec: zephyr_discovery_types.MeshSpec{
							MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-3"},
						Spec: zephyr_discovery_types.MeshSpec{
							MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
				}
				// contents don't matter
				clusterNameToSnapshot := map[string]*snapshot.TranslatedSnapshot{
					"cluster1": {},
					"cluster2": {},
				}

				meshClient.EXPECT().
					ListMesh(ctx).
					Return(&zephyr_discovery.MeshList{
						Items: []zephyr_discovery.Mesh{*knownMeshes[0], *knownMeshes[1], *knownMeshes[2]},
					}, nil)
				meshServiceClient.EXPECT().
					ListMeshService(ctx).
					Return(&zephyr_discovery.MeshServiceList{}, nil)
				snapshotReconciler.EXPECT().
					ReconcileAllSnapshots(ctx, clusterNameToSnapshot).
					Return(nil)

				err := reconciler.Reconcile(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("we have traffic targets (services or workloads) on those meshes", func() {
			FIt("generates the correct resources to be reconciled", func() {
				meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
				meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
				snapshotReconciler := mock_snapshot.NewMockTranslationSnapshotReconciler(ctrl)
				snapshotAccumulator := mock_snapshot.NewMockTranslationSnapshotAccumulator(ctrl)
				var snapshotAccumulatorGetter snapshot.TranslationSnapshotAccumulatorGetter = func(meshType zephyr_core_types.MeshType) (accumulator snapshot.TranslationSnapshotAccumulator, err error) {
					return snapshotAccumulator, nil
				}
				reconciler := translation_framework.NewTranslationReconciler(meshServiceClient, meshClient, snapshotAccumulatorGetter, snapshotReconciler)
				knownMeshes := []*zephyr_discovery.Mesh{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-1"},
						Spec: zephyr_discovery_types.MeshSpec{
							Cluster:  &zephyr_core_types.ResourceRef{Name: "cluster-1"},
							MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-with-no-services"},
						Spec: zephyr_discovery_types.MeshSpec{
							Cluster:  &zephyr_core_types.ResourceRef{Name: "cluster-2"},
							MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-3"},
						Spec: zephyr_discovery_types.MeshSpec{
							Cluster:  &zephyr_core_types.ResourceRef{Name: "cluster-3"},
							MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
				}
				meshServices := []*zephyr_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms1"},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(knownMeshes[0].ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms2"},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(knownMeshes[0].ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms3"},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(knownMeshes[2].ObjectMeta),
						},
					},
				}

				meshClient.EXPECT().
					ListMesh(ctx).
					Return(&zephyr_discovery.MeshList{
						Items: []zephyr_discovery.Mesh{*knownMeshes[0], *knownMeshes[1], *knownMeshes[2]},
					}, nil)
				meshServiceClient.EXPECT().
					ListMeshService(ctx).
					Return(&zephyr_discovery.MeshServiceList{
						Items: []zephyr_discovery.MeshService{*meshServices[0], *meshServices[1], *meshServices[2]},
					}, nil)
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[0], meshServices, knownMeshes[0]).
					DoAndReturn(func(
						snapshotInProgress *snapshot.TranslatedSnapshot,
						meshService *zephyr_discovery.MeshService,
						allMeshServices []*zephyr_discovery.MeshService,
						mesh *zephyr_discovery.Mesh,
					) error {
						snapshotInProgress.Istio = &snapshot.IstioSnapshot{
							DestinationRules: []*istio_networking.DestinationRule{{
								Host: "host-1",
							}},
						}
						return nil
					})
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[1], meshServices, knownMeshes[0]).
					DoAndReturn(func(
						snapshotInProgress *snapshot.TranslatedSnapshot,
						meshService *zephyr_discovery.MeshService,
						allMeshServices []*zephyr_discovery.MeshService,
						mesh *zephyr_discovery.Mesh,
					) error {
						snapshotInProgress.Istio.DestinationRules = append(snapshotInProgress.Istio.DestinationRules, &istio_networking.DestinationRule{
							Host: "host-2",
						})
						return nil
					})
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[2], meshServices, knownMeshes[2]).
					DoAndReturn(func(
						snapshotInProgress *snapshot.TranslatedSnapshot,
						meshService *zephyr_discovery.MeshService,
						allMeshServices []*zephyr_discovery.MeshService,
						mesh *zephyr_discovery.Mesh,
					) error {
						snapshotInProgress.Istio = &snapshot.IstioSnapshot{
							DestinationRules: []*istio_networking.DestinationRule{{
								Host: "host-3",
							}},
						}
						return nil
					})

				expectedClusterNameToSnapshot := translation_framework.NewClusterNameToSnapshot(knownMeshes)
				expectedClusterNameToSnapshot[translation_framework.ClusterKeyFromMesh(knownMeshes[0])].Istio = &snapshot.IstioSnapshot{
					DestinationRules: []*istio_networking.DestinationRule{{
						Host: "host-1",
					}, {
						Host: "host-2",
					}},
				}
				expectedClusterNameToSnapshot[translation_framework.ClusterKeyFromMesh(knownMeshes[2])].Istio = &snapshot.IstioSnapshot{
					DestinationRules: []*istio_networking.DestinationRule{{
						Host: "host-3",
					}},
				}

				snapshotReconciler.EXPECT().
					ReconcileAllSnapshots(ctx, ClusterSnapshotMatcher(expectedClusterNameToSnapshot)).
					Return(nil)

				err := reconciler.Reconcile(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

func ClusterSnapshotMatcher(cnts snapshot.ClusterNameToSnapshot) gomock.Matcher {
	return gomock.GotFormatterAdapter(
		gomock.GotFormatterFunc(func(i interface{}) string {
			return spew.Sdump(i)
		}), &clusterSnapshotMatcher{
			this: cnts,
		},
	)
}

type clusterSnapshotMatcher struct {
	this snapshot.ClusterNameToSnapshot
}

// Matches returns whether x is a match.
func (c *clusterSnapshotMatcher) Matches(x interface{}) bool {
	if other, ok := x.(snapshot.ClusterNameToSnapshot); ok {
		if len(other) != len(c.this) {
			return false
		}
		for k, v := range other {
			if v2, ok := c.this[k]; ok {
				if !reflect.DeepEqual(v, v2) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}

// String describes what the matcher matches.
func (c *clusterSnapshotMatcher) String() string {
	return spew.Sdump(c.this)
}
