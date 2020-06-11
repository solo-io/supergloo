package mesh_translation

import (
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type MeshTranslatorFactory struct {
	istio mesh_translation.IstioTranslator
}

func NewMeshTranslatorFactory(istio mesh_translation.IstioTranslator) {
	return &MeshTranslatorFactory{
		istio: istio,
	}
}

func (t *MeshTyper) MeshTypeToAccumulator(meshType zephyr_core_types.MeshType) (snapshot.TranslationSnapshotAccumulator, error) {
	switch meshType {
	case zephyr_core_types.MeshType_ISTIO1_5:
		fallthrough
	case zephyr_core_types.MeshType_ISTIO1_6:
		return t.istio, nil
	default:
		return nil, errors.Errorf("mesh not supported")
	}
}

func (t *MeshTyper) MeshTypeToTranslationValidator(meshType zephyr_core_types.MeshType) (mesh_translation.TranslationValidator, error) {
	switch meshType {
	case zephyr_core_types.MeshType_ISTIO1_5:
		fallthrough
	case zephyr_core_types.MeshType_ISTIO1_6:
		return t.istio, nil
	default:
		return nil, errors.Errorf("mesh not supported")

	}
}
