package vm_validation

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	VirtualMeshFinder is a higher-level client aimed at simplifying the finding of meshes on VirtualMeshes
*/
type VirtualMeshFinder interface {
	GetMeshesForVirtualMesh(ctx context.Context, vm *smh_networking.VirtualMesh) ([]*smh_discovery.Mesh, error)
}
