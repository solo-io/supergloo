package translation_framework

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
)

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
	logger := contextutils.LoggerFrom(ctx)
	logger.Debug("Running iteration of traffic policy translator")

	meshList, err := t.meshClient.ListMesh(ctx)
	if err != nil {
		return err
	}

	// need to populate this map from our known meshes, rather than the mesh services we know about
	// in the case of deleting all mesh services but the mesh remains, we want to be sure to reconcile properly
	var knownMeshes []*zephyr_discovery.Mesh
	meshIdToMesh := map[string]*zephyr_discovery.Mesh{}
	for _, meshIter := range meshList.Items {
		mesh := meshIter
		knownMeshes = append(knownMeshes, &mesh)
		meshIdToMesh[clients.ToUniqueSingleClusterString(mesh.ObjectMeta)] = &mesh
	}

	if len(knownMeshes) == 0 {
		return nil
	}

	clusterNameToSnapshot := t.snapshotReconciler.InitializeClusterNameToSnapshot(knownMeshes)

	meshServiceList, err := t.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return err
	}

	var allMeshServices []*zephyr_discovery.MeshService
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		allMeshServices = append(allMeshServices, &meshService)
	}

	for _, meshServiceIter := range meshServiceList.Items {
		meshService := meshServiceIter

		meshId := clients.ToUniqueSingleClusterString(clients.ResourceRefToObjectMeta(meshService.Spec.GetMesh()))
		mesh, ok := meshIdToMesh[meshId]
		if !ok {
			return eris.Errorf("Got a mesh service %s.%s belonging to a mesh %s.%s that does not exist", meshService.GetName(), meshService.GetNamespace(), mesh.GetName(), mesh.GetNamespace())
		}

		clusterName := mesh.Spec.GetCluster().GetName()
		meshType, err := enum_conversion.MeshToMeshType(mesh)
		if err != nil {
			return err
		}

		snapshotAccumulator, err := t.translationSnapshotBuilderGetter(meshType)
		if err != nil {
			return err
		}

		// run one round of translation just for this service, accumulating the results into our map
		err = snapshotAccumulator.AccumulateFromTranslation(
			clusterNameToSnapshot[clusterName],
			&meshService,
			allMeshServices,
			mesh,
		)
		if err != nil {
			return err
		}

	}

	// reconcile everything at once
	return t.snapshotReconciler.ReconcileAllSnapshots(ctx, clusterNameToSnapshot)
}
