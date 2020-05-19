package translation_framework

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
)

func NewTranslationReconciler(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
	translationSnapshotBuilderGetter TranslationSnapshotAccumulatorGetter,
	snapshotReconciler TranslationSnapshotReconciler,
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
	translationSnapshotBuilderGetter TranslationSnapshotAccumulatorGetter
	snapshotReconciler               TranslationSnapshotReconciler
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

	// need to populate this map from our known clusters, rather than the mesh services we know about
	// in the case of deleting all mesh services but the mesh remains, we want to be sure to reconcile properly
	// an entry
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
			mesh,
		)
		if err != nil {
			return err
		}

	}

	// reconcile everything at once
	return t.snapshotReconciler.ReconcileAllSnapshots(ctx, clusterNameToSnapshot)
}

//		err = t.accumulateResourcesForService(
//			logger,
//			&meshService,
//			mesh,
//			meshType,
//			clusterNameToResources,
//		)
//		if err != nil {
//			return err
//		}
//	}
//
//	return t.reconcileOutputResources(ctx, clusterNameToResources)
//}
//
//func (t *translationReconciler) reconcileOutputResources(ctx context.Context, clusterNameToResources map[string]*translatedSnapshot) error {
//	for clusterName, resourcesToReconcile := range clusterNameToResources {
//		client, err := t.clientGetter.GetClientForCluster(ctx, clusterName)
//		if err != nil {
//			return err
//		}
//
//		if resourcesToReconcile.istio != nil {
//			virtualServiceReconciler, err := t.virtualServiceReconcilerBuilder.
//				WithClient(client).
//				ScopedToLabels(resourcesToReconcile.istio.translationLabels).
//				Build()
//			if err != nil {
//				return err
//			}
//
//			err = virtualServiceReconciler.Reconcile(ctx, resourcesToReconcile.istio.VirtualServices)
//			if err != nil {
//				return err
//			}
//
//			destinationRuleReconciler, err := t.destinationRuleReconcilerBuilder.
//				WithClient(client).
//				ScopedToLabels(resourcesToReconcile.istio.translationLabels).
//				Build()
//			if err != nil {
//				return err
//			}
//
//			err = destinationRuleReconciler.Reconcile(ctx, resourcesToReconcile.istio.DestinationRules)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (t *translationReconciler) accumulateResourcesForService(
//	logger *zap.SugaredLogger,
//	meshService *zephyr_discovery.MeshService,
//	mesh *zephyr_discovery.Mesh,
//	meshType zephyr_core_types.MeshType,
//	clusterNameToResources map[string]*translatedSnapshot,
//) error {
//	logger.Debugf("Translating for mesh service %s.%s", meshService.GetName(), meshService.GetNamespace())
//
//	clusterName := mesh.Spec.GetCluster().GetName()
//
//	switch meshType {
//	case zephyr_core_types.MeshType_ISTIO:
//		if _, ok := clusterNameToResources[clusterName]; !ok {
//			clusterNameToResources[clusterName] = &translatedSnapshot{
//				istio: &istioSnapshot{
//					translationLabels: t.istioTranslator.GetTranslationLabels(),
//				},
//			}
//		}
//
//		output, translationErr := t.istioTranslator.Translate(meshService, mesh, meshService.Status.ValidatedTrafficPolicies)
//		if len(translationErr) > 0 {
//			return eris.Errorf("Translation errors occurred in translation reconciler; this is unexpected: %+v", translationErr)
//		}
//
//		existingResources := clusterNameToResources[clusterName]
//		existingResources.istio.DestinationRules = append(existingResources.istio.DestinationRules, output.DestinationRules...)
//		existingResources.istio.VirtualServices = append(existingResources.istio.VirtualServices, output.VirtualServices...)
//	default:
//		return eris.Errorf("Traffic policy translation is unsupported for mesh type %s", meshType.String())
//	}
//
//	return nil
//}
