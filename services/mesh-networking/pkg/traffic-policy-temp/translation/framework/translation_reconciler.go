package translation_framework

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	"k8s.io/apimachinery/pkg/types"
)

type TranslationProcessor interface {
	Process(ctx context.Context) (snapshot.ClusterNameToSnapshot, error)
}

func NewTranslationReconciler(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
	translationSnapshotBuilderGetter snapshot.TranslationSnapshotAccumulatorGetter,
	snapshotReconciler snapshot.TranslationSnapshotReconciler,
) reconciliation.Reconciler {
	return &translationReconciler{
		meshServiceClient:                meshServiceClient,
		meshClient:                       meshClient,
		translationSnapshotBuilderGetter: translationSnapshotBuilderGetter,
		snapshotReconciler:               snapshotReconciler,
	}
}

type translationReconciler struct {
	meshServiceClient                zephyr_discovery.MeshServiceClient
	meshClient                       zephyr_discovery.MeshClient
	translationSnapshotBuilderGetter snapshot.TranslationSnapshotAccumulatorGetter
	snapshotReconciler               snapshot.TranslationSnapshotReconciler
}

func (*translationReconciler) GetName() string {
	return "traffic-policy-translation-reconciler"
}

func (t *translationReconciler) Reconcile(ctx context.Context) error {
	processor := translationProcessor{
		meshServiceReader:                t.meshServiceClient,
		meshReader:                       t.meshClient,
		translationSnapshotBuilderGetter: t.translationSnapshotBuilderGetter,
	}
	clusterNameToSnapshot, err := processor.Process(ctx)
	if err != nil {
		return err
	}
	if clusterNameToSnapshot == nil {
		return nil
	}
	// reconcile everything at once
	return t.snapshotReconciler.ReconcileAllSnapshots(ctx, clusterNameToSnapshot)

}

type translationProcessor struct {
	meshServiceReader                zephyr_discovery.MeshServiceReader
	meshReader                       zephyr_discovery.MeshReader
	translationSnapshotBuilderGetter snapshot.TranslationSnapshotAccumulatorGetter
}

func ClusterKeyFromMesh(mesh *zephyr_discovery.Mesh) types.NamespacedName {
	return types.NamespacedName{
		Name:      mesh.Spec.Cluster.Name,
		Namespace: mesh.Spec.Cluster.Namespace,
	}
}

// return a map that's pre-populated for every cluster name referenced in the meshes
// we still want to run reconciliation for clusters where there are no mesh services
func NewClusterNameToSnapshot(knownMeshes []*zephyr_discovery.Mesh) snapshot.ClusterNameToSnapshot {
	m := snapshot.ClusterNameToSnapshot{}
	for _, mesh := range knownMeshes {
		m[ClusterKeyFromMesh(mesh)] = &snapshot.TranslatedSnapshot{}
	}
	return m
}

func (t *translationProcessor) Process(ctx context.Context) (snapshot.ClusterNameToSnapshot, error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Debug("Running iteration of traffic policy translator")

	meshList, err := t.meshReader.ListMesh(ctx)
	if err != nil {
		return nil, err
	}

	// need to populate this map from our known meshes, rather than the mesh services we know about
	// in the case of deleting all mesh services but the mesh remains, we want to be sure to reconcile properly
	var knownMeshes []*zephyr_discovery.Mesh
	meshIdToMesh := map[string]*zephyr_discovery.Mesh{}
	for _, meshIter := range meshList.Items {
		mesh := meshIter
		knownMeshes = append(knownMeshes, &mesh)
		meshIdToMesh[selection.ToUniqueSingleClusterString(mesh.ObjectMeta)] = &mesh
	}

	if len(knownMeshes) == 0 {
		return nil, nil
	}

	clusterNameToSnapshot := NewClusterNameToSnapshot(knownMeshes)

	meshServiceList, err := t.meshServiceReader.ListMeshService(ctx)
	if err != nil {
		return nil, err
	}

	var allMeshServices []*zephyr_discovery.MeshService
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		allMeshServices = append(allMeshServices, &meshService)
	}

	for _, meshServiceIter := range meshServiceList.Items {
		meshService := meshServiceIter

		meshId := selection.ToUniqueSingleClusterString(selection.ResourceRefToObjectMeta(meshService.Spec.GetMesh()))
		mesh, ok := meshIdToMesh[meshId]
		if !ok {
			return nil, eris.Errorf("Got a mesh service %s.%s belonging to a mesh %s.%s that does not exist", meshService.GetName(), meshService.GetNamespace(), mesh.GetName(), mesh.GetNamespace())
		}

		meshType, err := metadata.MeshToMeshType(mesh)
		if err != nil {
			return nil, err
		}

		snapshotAccumulator, err := t.translationSnapshotBuilderGetter(meshType)
		if err != nil {
			return nil, err
		}

		// run one round of translation just for this service, accumulating the results into our map
		err = snapshotAccumulator.AccumulateFromTranslation(
			clusterNameToSnapshot[ClusterKeyFromMesh(mesh)],
			&meshService,
			allMeshServices,
			mesh,
		)
		if err != nil {
			return nil, err
		}

	}

	return clusterNameToSnapshot, nil
}
