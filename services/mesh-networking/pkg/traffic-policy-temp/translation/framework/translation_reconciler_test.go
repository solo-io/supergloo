package translation_framework_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	translation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	mock_snapshot "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot/mocks"
	mock_smh_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	"istio.io/api/networking/v1alpha3"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TranslationReconciler", func() {
	var (
		ctx                       context.Context
		ctrl                      *gomock.Controller
		meshClient                *mock_smh_discovery_clients.MockMeshClient
		snapshotAccumulator       *mock_snapshot.MockTranslationSnapshotAccumulator
		snapshotAccumulatorGetter snapshot.TranslationSnapshotAccumulatorGetter = func(meshType smh_core_types.MeshType) (accumulator snapshot.TranslationSnapshotAccumulator, err error) {
			return snapshotAccumulator, nil
		}
		processor translation_framework.TranslationProcessor
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
		meshClient = mock_smh_discovery_clients.NewMockMeshClient(ctrl)
		snapshotAccumulator = mock_snapshot.NewMockTranslationSnapshotAccumulator(ctrl)
		processor = translation_framework.NewTranslationProcessor(meshClient, snapshotAccumulatorGetter)

	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("no resources exist", func() {
		It("does nothing", func() {
			meshClient.EXPECT().
				ListMesh(ctx).
				Return(&smh_discovery.MeshList{}, nil)

			snapshot, err := processor.Process(ctx, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapshot).To(BeEmpty())
		})
	})

	When("we have meshes", func() {
		When("we have no traffic targets (services or workloads) on those meshes", func() {
			It("still runs the output reconciliation", func() {
				knownMeshes := []*smh_discovery.Mesh{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-1"},
						Spec: smh_discovery_types.MeshSpec{
							MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
							Cluster: &smh_core_types.ResourceRef{
								Name: "cluster1",
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-2"},
						Spec: smh_discovery_types.MeshSpec{
							MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
							Cluster: &smh_core_types.ResourceRef{
								Name: "cluster2",
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-3"},
						Spec: smh_discovery_types.MeshSpec{
							MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
							Cluster: &smh_core_types.ResourceRef{
								Name: "cluster2",
							},
						},
					},
				}
				// contents don't matter
				clusterNameToSnapshot := snapshot.ClusterNameToSnapshot{
					translation_framework.ClusterKeyFromMesh(knownMeshes[0]): {},
					translation_framework.ClusterKeyFromMesh(knownMeshes[1]): {},
				}

				meshClient.EXPECT().
					ListMesh(ctx).
					Return(&smh_discovery.MeshList{
						Items: []smh_discovery.Mesh{*knownMeshes[0], *knownMeshes[1], *knownMeshes[2]},
					}, nil)

				outSnapshot, err := processor.Process(ctx, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(outSnapshot).To(Equal(clusterNameToSnapshot))
			})
		})

		When("we have traffic targets (services or workloads) on those meshes", func() {
			It("generates the correct resources to be reconciled", func() {
				knownMeshes := []*smh_discovery.Mesh{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-1"},
						Spec: smh_discovery_types.MeshSpec{
							Cluster:  &smh_core_types.ResourceRef{Name: "cluster-1"},
							MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-with-no-services"},
						Spec: smh_discovery_types.MeshSpec{
							Cluster:  &smh_core_types.ResourceRef{Name: "cluster-2"},
							MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-3"},
						Spec: smh_discovery_types.MeshSpec{
							Cluster:  &smh_core_types.ResourceRef{Name: "cluster-3"},
							MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						},
					},
				}
				meshServices := []*smh_discovery.MeshService{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms1"},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(knownMeshes[0].ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms2"},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(knownMeshes[0].ObjectMeta),
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms3"},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(knownMeshes[2].ObjectMeta),
						},
					},
				}

				meshClient.EXPECT().
					ListMesh(ctx).
					Return(&smh_discovery.MeshList{
						Items: []smh_discovery.Mesh{*knownMeshes[0], *knownMeshes[1], *knownMeshes[2]},
					}, nil)
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[0], meshServices, knownMeshes[0]).
					DoAndReturn(func(
						snapshotInProgress *snapshot.TranslatedSnapshot,
						meshService *smh_discovery.MeshService,
						allMeshServices []*smh_discovery.MeshService,
						mesh *smh_discovery.Mesh,
					) error {
						snapshotInProgress.Istio = &snapshot.IstioSnapshot{
							DestinationRules: []*istio_networking.DestinationRule{{
								Spec: v1alpha3.DestinationRule{
									Host: "host-1",
								},
							}},
						}
						return nil
					})
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[1], meshServices, knownMeshes[0]).
					DoAndReturn(func(
						snapshotInProgress *snapshot.TranslatedSnapshot,
						meshService *smh_discovery.MeshService,
						allMeshServices []*smh_discovery.MeshService,
						mesh *smh_discovery.Mesh,
					) error {
						snapshotInProgress.Istio.DestinationRules = append(snapshotInProgress.Istio.DestinationRules, &istio_networking.DestinationRule{
							Spec: v1alpha3.DestinationRule{
								Host: "host-2",
							},
						})
						return nil
					})
				snapshotAccumulator.EXPECT().
					AccumulateFromTranslation(gomock.Any(), meshServices[2], meshServices, knownMeshes[2]).
					DoAndReturn(func(
						snapshotInProgress *snapshot.TranslatedSnapshot,
						meshService *smh_discovery.MeshService,
						allMeshServices []*smh_discovery.MeshService,
						mesh *smh_discovery.Mesh,
					) error {
						snapshotInProgress.Istio = &snapshot.IstioSnapshot{
							DestinationRules: []*istio_networking.DestinationRule{{
								Spec: v1alpha3.DestinationRule{
									Host: "host-3",
								},
							}},
						}
						return nil
					})

				expectedClusterNameToSnapshot := translation_framework.NewClusterNameToSnapshot(knownMeshes)
				expectedClusterNameToSnapshot[translation_framework.ClusterKeyFromMesh(knownMeshes[0])].Istio = &snapshot.IstioSnapshot{
					DestinationRules: []*istio_networking.DestinationRule{{
						Spec: v1alpha3.DestinationRule{
							Host: "host-1",
						},
					}, {
						Spec: v1alpha3.DestinationRule{
							Host: "host-2",
						},
					}},
				}
				expectedClusterNameToSnapshot[translation_framework.ClusterKeyFromMesh(knownMeshes[2])].Istio = &snapshot.IstioSnapshot{
					DestinationRules: []*istio_networking.DestinationRule{{
						Spec: v1alpha3.DestinationRule{
							Host: "host-3",
						},
					}},
				}

				outSnapshot, err := processor.Process(ctx, meshServices)
				Expect(err).NotTo(HaveOccurred())
				Expect(outSnapshot).To(Equal(expectedClusterNameToSnapshot))

			})
		})
	})
})
