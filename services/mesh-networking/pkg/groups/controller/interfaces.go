package group_controller

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go
/*
	The MeshGroupValidator is meant to check the validity of any MeshGroup resource, and return the updated status for
	said resource. The properties it is testing are the ones which cannot be tested by a simple JSON schema check, as
	in the future this will be done using a schema in the CRD spec.
*/
type MeshGroupValidator interface {
	/*
		Validate takes as arguments the ctx of the event, as well as the mesh group needing to be validated.

		The return states are as follows:

		1. MeshGroupStatus_PROCESSING_ERROR, err: This means that there was an error trying to determine the
		validity of the mesh group, and the event should be requeued.
		2. MeshGroupStatus_INVALID, err: This means that the mesh group was processed properly, and there was
		found to be an error with the configuration. This event should not be requeued.
		3. MeshGroupStatus_VALID, nil: This means that the mesh group was processed properly, and the mesh group
		is valid
	*/
	Validate(ctx context.Context, mg *v1alpha1.MeshGroup) (types.MeshGroupStatus, error)
}
