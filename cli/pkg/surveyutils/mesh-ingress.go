package surveyutils

import (
	"context"

	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
)

func SurveyMeshIngresses(ctx context.Context) ([]*core.ResourceRef, error) {
	// collect Meshes list
	meshIngressClient := clients.MustMeshIngressClient()
	meshIngresses, err := meshIngressClient.List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	var selected []*core.ResourceRef
	for {
		meshIngress, err := surveyResources("meshingresses", "add a mesh ingress (choose <done> to finish): ", "<done>", meshIngresses.AsResources())
		if err != nil {
			return nil, err
		}
		// the user chose
		if meshIngress.Namespace == "" {
			return selected, nil
		}
		selected = append(selected, &meshIngress)
	}
}

func SurveyMeshIngress(ctx context.Context) (core.ResourceRef, error) {
	meshIngresses, err := clients.MustMeshIngressClient().List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return core.ResourceRef{}, err
	}
	return surveyResources("meshingresses", "select a mesh ingress: ", "", meshIngresses.AsResources())
}
