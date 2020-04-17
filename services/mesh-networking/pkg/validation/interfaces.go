package vm_validation

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	VirtualMeshFinder is a higher-level client aimed at simplifying the finding of meshes on VirtualMeshes
*/
type VirtualMeshFinder interface {
	GetMeshesForVirtualMesh(ctx context.Context, vm *zephyr_networking.VirtualMesh) ([]*zephyr_discovery.Mesh, error)
}
