package translation_framework

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework/snapshot"
	"k8s.io/apimachinery/pkg/types"
)

//go:generate mockgen -source ./translation_processor.go -destination ./mocks/mock_translation_processor.go

type TranslationProcessor interface {
	Process(ctx context.Context, allMeshServices []*smh_discovery.MeshService) (snapshot.ClusterNameToSnapshot, error)
}

func NewTranslationProcessor(
	meshReader smh_discovery.MeshReader,
	translationSnapshotBuilderGetter snapshot.TranslationSnapshotAccumulatorGetter,
) TranslationProcessor {
	return &translationProcessor{
		meshReader:                       meshReader,
		translationSnapshotBuilderGetter: translationSnapshotBuilderGetter,
	}
}

type translationProcessor struct {
	meshReader                       smh_discovery.MeshReader
	translationSnapshotBuilderGetter snapshot.TranslationSnapshotAccumulatorGetter
}

func ClusterKeyFromMesh(mesh *smh_discovery.Mesh) types.NamespacedName {
	return types.NamespacedName{
		Name:      mesh.Spec.GetCluster().GetName(),
		Namespace: mesh.Spec.GetCluster().GetNamespace(),
	}
}

// return a map that's pre-populated for every cluster name referenced in the meshes
// we still want to run reconciliation for clusters where there are no mesh services
func NewClusterNameToSnapshot(knownMeshes []*smh_discovery.Mesh) snapshot.ClusterNameToSnapshot {
	m := snapshot.ClusterNameToSnapshot{}
	for _, mesh := range knownMeshes {
		if !meshIsSupported(mesh) {
			continue
		}
		m[ClusterKeyFromMesh(mesh)] = &snapshot.TranslatedSnapshot{}
	}
	return m
}

func (t *translationProcessor) Process(ctx context.Context, allMeshServices []*smh_discovery.MeshService) (snapshot.ClusterNameToSnapshot, error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Debug("Running iteration of traffic policy translator")

	meshList, err := t.meshReader.ListMesh(ctx)
	if err != nil {
		return nil, err
	}

	// need to populate this map from our known meshes, rather than the mesh services we know about
	// in the case of deleting all mesh services but the mesh remains, we want to be sure to reconcile properly
	var knownMeshes []*smh_discovery.Mesh
	meshIdToMesh := map[string]*smh_discovery.Mesh{}
	for _, meshIter := range meshList.Items {
		mesh := meshIter
		knownMeshes = append(knownMeshes, &mesh)
		meshIdToMesh[selection.ToUniqueSingleClusterString(mesh.ObjectMeta)] = &mesh
	}

	if len(knownMeshes) == 0 {
		return nil, nil
	}

	clusterNameToSnapshot := NewClusterNameToSnapshot(knownMeshes)

	var multierr error
	for _, meshService := range allMeshServices {
		err := t.processService(ctx, meshService, allMeshServices, meshIdToMesh, clusterNameToSnapshot)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}

	logger.Debugw("translation processor", "snapshot", clusterNameToSnapshot)
	return clusterNameToSnapshot, multierr
}

func (t *translationProcessor) processService(
	ctx context.Context,
	meshService *smh_discovery.MeshService,
	allMeshServices []*smh_discovery.MeshService,
	meshIdToMesh map[string]*smh_discovery.Mesh,
	clusterNameToSnapshot snapshot.ClusterNameToSnapshot,
) error {
	logger := contextutils.LoggerFrom(ctx)

	meshId := selection.ToUniqueSingleClusterString(selection.ResourceRefToObjectMeta(meshService.Spec.GetMesh()))
	mesh, ok := meshIdToMesh[meshId]
	if !ok {
		meshRef := meshService.Spec.GetMesh()
		return eris.Errorf("Got a mesh service %s.%s belonging to a mesh %s.%s that does not exist",
			meshService.GetName(),
			meshService.GetNamespace(),
			meshRef.GetName(),
			meshRef.GetNamespace(),
		)
	}

	meshType, err := metadata.MeshToMeshType(mesh)
	if err != nil {
		// TODO: once all the code is in the new model, add error:
		// return err
		return nil
	}

	snapshotAccumulator, err := t.translationSnapshotBuilderGetter(meshType)
	if err != nil {
		return nil
	}

	// we run translation even if the service has translation errors - as we might want to
	// partially translate what we can.
	meshKey := ClusterKeyFromMesh(mesh)
	snapshot := clusterNameToSnapshot[meshKey]
	logger.Debugw("accumulate from translation", "meshService", meshService, "mesh", mesh, "snapshot", snapshot, "meshKey", meshKey)
	err = snapshotAccumulator.AccumulateFromTranslation(
		ctx,
		snapshot,
		meshService,
		allMeshServices,
		mesh,
	)
	return err
}

func meshIsSupported(mesh *smh_discovery.Mesh) bool {
	meshType, err := metadata.MeshToMeshType(mesh)
	if err != nil {
		return false
	}
	switch meshType {
	case smh_core_types.MeshType_ISTIO1_5:
		fallthrough
	case smh_core_types.MeshType_ISTIO1_6:
		return true
	default:
		return false
	}
}
