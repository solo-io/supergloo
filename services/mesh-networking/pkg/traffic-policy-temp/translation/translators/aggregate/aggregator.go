package mesh_translation

import (
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators/istio"
	"github.com/solo-io/solo-kit/pkg/errors"
)

func MeshTypeToAccumulator(meshType zephyr_core_types.MeshType) (snapshot.TranslationSnapshotAccumulator, error) {
	switch meshType {
	case zephyr_core_types.MeshType_ISTIO1_5:
		fallthrough
	case zephyr_core_types.MeshType_ISTIO1_6:
		return istio_translator.NewIstioTrafficPolicyTranslator(selection.NewBaseResourceSelector()), nil
	default:
		return nil, errors.Errorf("mesh not supported")

	}
}
func MeshTypeToTranslationValidator(meshType zephyr_core_types.MeshType) (mesh_translation.TranslationValidator, error) {
	switch meshType {
	case zephyr_core_types.MeshType_ISTIO1_5:
		fallthrough
	case zephyr_core_types.MeshType_ISTIO1_6:
		return istio_translator.NewIstioTrafficPolicyTranslator(selection.NewBaseResourceSelector()), nil
	default:
		return nil, errors.Errorf("mesh not supported")

	}
}
