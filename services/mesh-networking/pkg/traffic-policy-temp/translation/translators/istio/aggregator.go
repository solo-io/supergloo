package istio_translator

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
)

type IstioTranslationSnapshotAccumulator struct {
}

// mutate the translated snapshot, adding the translation results in where appropriate
func (i *IstioTranslationSnapshotAccumulator) AccumulateFromTranslation(
	snapshotInProgress *snapshot.TranslatedSnapshot,
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	translator mesh_translation.IstioTranslator,
) error {
	if snapshotInProgress.Istio == nil {
		snapshotInProgress.Istio = &snapshot.IstioSnapshot{}
	}

	out, errs := translator.Translate(meshService, allMeshServices, mesh, meshService.Status.ValidatedTrafficPolicies)

	snapshotInProgress.Istio.DestinationRules = append(snapshotInProgress.Istio.DestinationRules, out.DestinationRules...)
	snapshotInProgress.Istio.VirtualServices = append(snapshotInProgress.Istio.VirtualServices, out.VirtualServices...)
	// assuming all went well in the previous stages, there should be no errors
	if len(errs) != 0 {
		// ?? panic ??
	}

	return nil
}
