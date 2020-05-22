package snapshot

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	istio_networking "istio.io/api/networking/v1alpha3"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// one of these will exist per mesh translation implementation
type TranslationSnapshotAccumulator interface {
	// mutate the translated snapshot, adding the translation results in where appropriate
	AccumulateFromTranslation(
		snapshotInProgress *TranslatedSnapshot,
		meshService *zephyr_discovery.MeshService,
		allMeshServices []*zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
	) error
}

type TranslationSnapshotReconciler interface {
	// return a map that's pre-populated for every cluster name referenced in the meshes
	// we still want to run reconciliation for clusters where there are no mesh services
	InitializeClusterNameToSnapshot(knownMeshes []*zephyr_discovery.Mesh) map[string]*TranslatedSnapshot
	ReconcileAllSnapshots(ctx context.Context, clusterNameToSnapshot map[string]*TranslatedSnapshot) error
}

type TranslationSnapshotAccumulatorGetter func(meshType zephyr_core_types.MeshType) (TranslationSnapshotAccumulator, error)

// a non-nil mesh field means that we need to run reconciliation on translated resources in that mesh
type TranslatedSnapshot struct {
	Istio *IstioSnapshot
}

type IstioSnapshot struct {
	TranslationLabels map[string]string
	DestinationRules  []*istio_networking.DestinationRule
	VirtualServices   []*istio_networking.VirtualService
}
