package translation_framework

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
)

func NewTranslationReconciler(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
) reconciliation.Reconciler {
	return &translationReconciler{
		meshServiceClient: meshServiceClient,
		meshClient:        meshClient,
	}
}

type translationReconciler struct {
	meshServiceClient zephyr_discovery.MeshServiceClient
	meshClient        zephyr_discovery.MeshClient
	istioTranslator   mesh_translation.IstioTranslator
}

func (*translationReconciler) GetName() string {
	return "traffic-policy-translation-reconciler"
}

func (t *translationReconciler) Reconcile(ctx context.Context) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Debug("Running iteration of traffic policy translator")

	meshServiceList, err := t.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return err
	}

	for _, meshServiceIter := range meshServiceList.Items {
		logger.Debugf("Translating for mesh service %s.%s", meshServiceIter.GetName(), meshServiceIter.GetNamespace())

		meshService := meshServiceIter

		mesh, err := t.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
		if err != nil {
			return err
		}

		meshType, err := enum_conversion.MeshToMeshType(mesh)
		if err != nil {
			return err
		}

		switch meshType {
		case zephyr_core_types.MeshType_ISTIO:
			output, translationErr := t.istioTranslator.Translate(&meshService, mesh, meshService.Status.ValidatedTrafficPolicies)
			if len(translationErr) > 0 {
				return eris.Errorf("Translation errors occurred in translation reconciler; this is unexpected: %+v", translationErr)
			}

		}
	}
}
