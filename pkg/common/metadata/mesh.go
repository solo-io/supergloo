package metadata

import (
	"fmt"
	"strings"

	"github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
)

func BuildMeshName(meshType types.MeshType, namespace string, cluster string) string {
	return fmt.Sprintf("%s-%s-%s", strings.ReplaceAll(strings.ToLower(meshType.String()), "_", "."), namespace, cluster)
}
