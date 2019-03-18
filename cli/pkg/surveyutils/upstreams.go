package surveyutils

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
)

func SurveyUpstreams(ctx context.Context) ([]core.ResourceRef, error) {
	// collect secrets list
	usClient := helpers.MustUpstreamClient()
	upstreams, err := usClient.List("", clients.ListOpts{Ctx: ctx})
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
