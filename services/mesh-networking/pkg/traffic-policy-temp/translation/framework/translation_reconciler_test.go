package translation_framework_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	translation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework"
	mock_translation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/mocks"
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
				snapshotReconciler := mock_translation_framework.NewMockTranslationSnapshotReconciler(ctrl)
				reconciler := translation_framework.NewTranslationReconciler(meshServiceClient, meshClient, nil, snapshotReconciler)
				knownMeshes := []*zephyr_discovery.Mesh{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-1"},
						Spec: zephyr_discovery_types.MeshSpec{
							MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-2"},
						Spec: zephyr_discovery_types.MeshSpec{
							MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-3"},
						Spec: zephyr_discovery_types.MeshSpec{
							MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
						},
					},
				}
				// contents don't matter
				clusterNameToSnapshot := map[string]*translation_framework.TranslatedSnapshot{
					"cluster1": nil,
					"cluster2": nil,
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
					InitializeClusterNameToSnapshot(knownMeshes).
					Return(clusterNameToSnapshot)
				snapshotReconciler.EXPECT().
					ReconcileAllSnapshots(ctx, clusterNameToSnapshot).
					Return(nil)

				err := reconciler.Reconcile(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("we have traffic targets (services or workloads) on those meshes", func() {
			It("generates the correct resources to be reconciled", func() {
				meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
				meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
				snapshotReconciler := mock_translation_framework.NewMockTranslationSnapshotReconciler(ctrl)
				snapshotAccumulator := mock_translation_framework.NewMockTranslationSnapshotAccumulator(ctrl)
				var snapshotAccumulatorGetter translation_framework.TranslationSnapshotAccumulatorGetter = func(meshType zephyr_core_types.MeshType) (accumulator translation_framework.TranslationSnapshotAccumulator, err error) {
					return snapshotAccumulator, nil
				}
				reconciler := translation_framework.NewTranslationReconciler(meshServiceClient, meshClient, snapshotAccumulatorGetter, snapshotReconciler)
				knownMeshes := []*zephyr_discovery.Mesh{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-1"},
						Spec: zephyr_discovery_types.MeshSpec{
							Cluster:  &zephyr_core_types.ResourceRef{Name: "cluster-1"},
							MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-with-no-services"},
						Spec: zephyr_discovery_types.MeshSpec{
							Cluster:  &zephyr_core_types.ResourceRef{Name: "cluster-2"},
							MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-3"},
						Spec: zephyr_discovery_types.MeshSpec{
							Cluster:  &zephyr_core_types.ResourceRef{Name: "cluster-3"},
							MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
						},
					},
				}
				meshServices := []*zephyr_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms1"},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(knownMeshes[0].ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms2"},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(knownMeshes[0].ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms3"},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(knownMeshes[2].ObjectMeta),
						},
					},
				}
				clusterNameToSnapshot := map[string]*translation_framework.TranslatedSnapshot{
					knownMeshes[0].Spec.Cluster.Name: {
						Istio: &translation_framework.IstioSnapshot{},
					},
					knownMeshes[1].Spec.Cluster.Name: {
						Istio: &translation_framework.IstioSnapshot{},
					},
					knownMeshes[2].Spec.Cluster.Name: {
						Istio: &translation_framework.IstioSnapshot{},
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
					AccumulateFromTranslation(gomock.Any(), meshServices[0], knownMeshes[0]).
					DoAndReturn(func(
						snapshotInProgress *translation_framework.TranslatedSnapshot,
						meshService *zephyr_discovery.MeshService,
						mesh *zephyr_discovery.Mesh,
					) error {
						snapshotInProgress.Istio = &translation_framework.IstioSnapshot{
							DestinationRules: []*istio_networking.DestinationRule{{
								Host: "host-1",
							}},
						}
						return nil
					})
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[1], knownMeshes[0]).
					DoAndReturn(func(
						snapshotInProgress *translation_framework.TranslatedSnapshot,
						meshService *zephyr_discovery.MeshService,
						mesh *zephyr_discovery.Mesh,
					) error {
						snapshotInProgress.Istio.DestinationRules = append(snapshotInProgress.Istio.DestinationRules, &istio_networking.DestinationRule{
							Host: "host-2",
						})
						return nil
					})
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[2], knownMeshes[2]).
					DoAndReturn(func(
						snapshotInProgress *translation_framework.TranslatedSnapshot,
						meshService *zephyr_discovery.MeshService,
						mesh *zephyr_discovery.Mesh,
					) error {
						snapshotInProgress.Istio = &translation_framework.IstioSnapshot{
							DestinationRules: []*istio_networking.DestinationRule{{
								Host: "host-3",
							}},
						}
						return nil
					})

				snapshotReconciler.EXPECT().
					InitializeClusterNameToSnapshot(knownMeshes).
					Return(clusterNameToSnapshot)
				snapshotReconciler.EXPECT().
					ReconcileAllSnapshots(ctx, clusterNameToSnapshot).
					Return(nil)

				err := reconciler.Reconcile(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
