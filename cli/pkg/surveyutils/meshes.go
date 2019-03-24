package surveyutils

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func SurveyMeshes(ctx context.Context) ([]*core.ResourceRef, error) {
	// collect Meshes list
	meshClient := clients.MustMeshClient()
	meshes, err := meshClient.List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	var selected []*core.ResourceRef
	for {
		mesh, err := surveyResources("meshes", "add a mesh (choose <done> to finish): ", "<done>", meshes.AsResources())
		if err != nil {
			return nil, err
		}
		// the user chose
		if mesh.Namespace == "" {
			return selected, nil
		}
		selected = append(selected, &mesh)
	}
}
