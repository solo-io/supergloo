package group_validation

import (
	"context"

	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	GroupMeshFinder is a higher-level client aimed at simplifying the finding of meshes on MeshGroups
*/
type GroupMeshFinder interface {
	GetMeshesForGroup(ctx context.Context, mg *networking_v1alpha1.MeshGroup) ([]*discoveryv1alpha1.Mesh, error)
}
