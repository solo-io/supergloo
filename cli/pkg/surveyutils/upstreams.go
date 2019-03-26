package surveyutils

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func SurveyUpstreams(ctx context.Context) ([]core.ResourceRef, error) {
	// collect secrets list
	usClient := clients.MustUpstreamClient()
	upstreams, err := usClient.List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	var selected []core.ResourceRef
	for {
		us, err := surveyResources("upstreams", "add an upstream (choose <done> to finish): ", "<done>", upstreams.AsResources())
		if err != nil {
			return nil, err
		}
		// the user chose
		if us.Namespace == "" {
			return selected, nil
		}
		selected = append(selected, us)
	}
}

func SurveyEditUpstream(ctx context.Context) (core.ResourceRef, error) {
	meshClient := clients.MustMeshClient()
	meshes, err := meshClient.List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return core.ResourceRef{}, err
	}

	mesh, err := surveyResources("meshes", "select a mesh", "", meshes.AsResources())
	if err != nil {
		return core.ResourceRef{}, err
	}

	return mesh, nil

}
