package mesh_translation_aggregate

import (
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework/snapshot"
	mesh_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type MeshTranslatorFactory struct {
	istio mesh_translation.IstioTranslator
}

func NewMeshTranslatorFactory(istio mesh_translation.IstioTranslator) *MeshTranslatorFactory {
	return &MeshTranslatorFactory{
		istio: istio,
	}
}

func (t *MeshTranslatorFactory) MeshTypeToAccumulator(meshType smh_core_types.MeshType) (snapshot.TranslationSnapshotAccumulator, error) {
	switch meshType {
	case smh_core_types.MeshType_ISTIO1_5:
		fallthrough
	case smh_core_types.MeshType_ISTIO1_6:
		return t.istio, nil
	default:
		return nil, errors.Errorf("mesh not supported")
	}
}

func (t *MeshTranslatorFactory) MeshTypeToTranslationValidator(meshType smh_core_types.MeshType) (mesh_translation.TranslationValidator, error) {
	switch meshType {
	case smh_core_types.MeshType_ISTIO1_5:
		fallthrough
	case smh_core_types.MeshType_ISTIO1_6:
		return t.istio, nil
	default:
		return nil, errors.Errorf("mesh not supported")

	}
}
