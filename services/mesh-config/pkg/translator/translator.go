package translator

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/common"
	"github.com/solo-io/mesh-projects/services/internal/config"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"go.uber.org/zap"
)

// This string will be used as a name for RbacConfig resources created by the operator.
// Istio's admission controller requires that this resource have the name "default"
const istioRbacConfigObjectName = "default"

var (
	statusInactive    = "no RBAC config specified"
	statusDisable     = "RBAC policies will not be evaluated"
	statusIsolate     = "RBAC is enabled and will deny-by-default. If no policies are defined (isolation mode requirement) no services will be allowed to communicate."
	statusConfigError = "Unknown config specified"
)

type Translator interface {
	Translate(ctx context.Context, in *v1.RbacSnapshot) (v1alpha1.ClusterRbacConfigList, error)
}

type clusterRbacTranslator struct {
	clientSet config.MeshConfigClientSet
}

func NewTranslator(clientSet config.MeshConfigClientSet) *clusterRbacTranslator {
	return &clusterRbacTranslator{
		clientSet: clientSet,
	}
}

func (t *clusterRbacTranslator) Translate(ctx context.Context, in *v1.RbacSnapshot) (v1alpha1.ClusterRbacConfigList, error) {
	var out v1alpha1.ClusterRbacConfigList
	for _, m := range in.Meshes {
		switch m.MeshType.(type) {
		case *v1.Mesh_Istio:
			nextRbacConfig := handleIstioRbac(ctx, m)
			if nextRbacConfig != nil {
				out = append(out, nextRbacConfig)
			}
		default:
			handleUnsupportedMeshes(m)
		}
	}
	return out, nil
}

func handleIstioRbac(ctx context.Context, istioMesh *v1.Mesh) *v1alpha1.ClusterRbacConfig {
	if istioMesh == nil {
		return nil
	}
	if istioMesh.GetRbac() == nil {
		return nil
	}
	switch istioMesh.Rbac.Mode.(type) {
	case *v1.RbacMode_Disable_:
		contextutils.LoggerFrom(ctx).Infow("syncing istio rbac to disable",
			zap.Any("mesh metadata", istioMesh.Metadata.String()))
		istioMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_OK,
			Message: statusDisable,
		}
		return &v1alpha1.ClusterRbacConfig{
			Metadata: newIstioRbacConfigMetadata(istioMesh),
			Mode:     v1alpha1.ClusterRbacConfig_OFF,
		}
	case *v1.RbacMode_Enable_:
		contextutils.LoggerFrom(ctx).Infow("syncing istio rbac to isolate",
			zap.Any("mesh metadata", istioMesh.Metadata.String()))
		istioMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_OK,
			Message: statusIsolate,
		}
		return &v1alpha1.ClusterRbacConfig{
			Metadata: newIstioRbacConfigMetadata(istioMesh),
			Mode:     v1alpha1.ClusterRbacConfig_ON,
		}
	default:
		istioMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_ERROR_CONFIG_NOT_ACCEPTED,
			Message: statusConfigError,
		}
		return nil
	}
}

// If a mesh is unsupported, it must not specify an RBAC config.
// Store the error in the mesh CRD if the user has provided an invalid config.
func handleUnsupportedMeshes(unsupportedInputMesh *v1.Mesh) *v1.Mesh {
	switch rbacMode := unsupportedInputMesh.GetRbac().GetMode().(type) {
	case *v1.RbacMode_Disable_:
		return unsupportedInputMesh
	default:
		unsupportedInputMesh.Rbac = &v1.RbacMode{
			Status: &v1.RbacStatus{
				Code:    v1.RbacStatusCode_ERROR_RBAC_MODE_NOT_SUPPORTED_BY_MESH,
				Message: fmt.Sprintf("cannot use rbac mode %v for this mesh type", rbacMode),
			},
		}
		return unsupportedInputMesh
	}
}

// helper for generating metadata for istio meshes
func newIstioRbacConfigMetadata(istioMesh *v1.Mesh) core.Metadata {
	return core.Metadata{
		Labels: common.OwnerLabels,
		// Must use the same cluster as the mesh as this is where it will be written
		Cluster: istioMesh.DiscoveryMetadata.Cluster,
		Name:    istioRbacConfigObjectName,
		// currently, we will install these resources in the InstallationNamespace
		// TODO - consider providing write namespace option
		Namespace: istioMesh.MeshType.(*v1.Mesh_Istio).Istio.InstallationNamespace,
	}
}
