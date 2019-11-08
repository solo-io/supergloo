package rbac

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"go.uber.org/zap"
)

// This string will be used as a name for RbacConfig resources created by the operator.
// Istio's admission controller requires that this resource have the name "default"
const istioRbacConfigObjectName = "default"

// helper for generating metadata for istio meshes
func newIstioRbacConfigMetadata(istioMesh *v1.Mesh) core.Metadata {
	return core.Metadata{
		// Must use the same cluster as the mesh as this is where it will be written
		Cluster: istioMesh.DiscoveryMetadata.Cluster,
		Name:    istioRbacConfigObjectName,
		// currently, we will install these resources in the InstallationNamespace
		// TODO - consider providing write namespace option
		Namespace: istioMesh.MeshType.(*v1.Mesh_Istio).Istio.InstallationNamespace,
	}
}

var (
	statusInactive    = "no RBAC config specified"
	statusDisable     = "RBAC policies will not be evaluated"
	statusIsolate     = "RBAC is enabled and will deny-by-default. If no policies are defined (isolation mode requirement) no services will be allowed to communicate."
	statusConfigError = "Unknown config specified"
)

func handleIstioRbac(ctx context.Context, istioMesh *v1.Mesh, rbacConfig *v1alpha1.RbacConfig) (*v1.Mesh, *v1alpha1.RbacConfig) {
	if istioMesh == nil {
		return nil, nil
	}
	outMesh := &v1.Mesh{}
	istioMesh.DeepCopyInto(outMesh)
	if istioMesh.Rbac == nil {
		return outMesh, nil
	}
	switch istioMesh.Rbac.Mode.(type) {
	case *v1.RbacMode_Unspecified_:
		outMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_OK,
			Message: statusInactive,
		}
		// return the existing RbacConfig (whether nil or active)
		return outMesh, rbacConfig
	case *v1.RbacMode_Disable_:
		contextutils.LoggerFrom(ctx).Infow("syncing istio rbac to disable",
			zap.Any("mesh metadata", istioMesh.Metadata.String()))
		outMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_OK,
			Message: statusDisable,
		}
		return outMesh, &v1alpha1.RbacConfig{
			Metadata: newIstioRbacConfigMetadata(istioMesh),
			Mode:     v1alpha1.RbacConfig_OFF,
		}
	case *v1.RbacMode_Enable_:
		contextutils.LoggerFrom(ctx).Infow("syncing istio rbac to isolate",
			zap.Any("mesh metadata", istioMesh.Metadata.String()))
		outMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_OK,
			Message: statusIsolate,
		}
		return outMesh, &v1alpha1.RbacConfig{
			Metadata: newIstioRbacConfigMetadata(istioMesh),
			Mode:     v1alpha1.RbacConfig_ON,
		}
	default:
		// this should not happen
		outMesh.Rbac.Status = &v1.RbacStatus{
			Code:    v1.RbacStatusCode_ERROR_CONFIG_NOT_ACCEPTED,
			Message: statusConfigError,
		}
		return outMesh, nil
	}
}

// If a mesh is unsupported, it must not specify an RBAC config.
// Store the error in the mesh CRD if the user has provided an invalid config.
func handleUnsupportedMeshes(unsupportedInputMesh *v1.Mesh) *v1.Mesh {
	if unsupportedInputMesh == nil {
		return nil
	}
	out := &v1.Mesh{}
	unsupportedInputMesh.DeepCopyInto(out)
	switch rbacMode := unsupportedInputMesh.Rbac.Mode.(type) {
	case *v1.RbacMode_Disable_:
		return out
	default:
		out.Rbac.Status.Code = v1.RbacStatusCode_ERROR_RBAC_MODE_NOT_SUPPORTED_BY_MESH
		out.Rbac.Status.Message = fmt.Sprintf("cannot use rbac mode %v for this mesh type", rbacMode)
		return out
	}
}
