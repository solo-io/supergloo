package surveyutils

import (
	"context"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/gloo/projects/gloo/cli/pkg/helpers"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func SelectUpstreams(ctx context.Context) ([]core.ResourceRef, error) {
	ups, err := helpers.MustUpstreamClient().List("", clients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	var selected []core.ResourceRef
	for {

	}
}

func getDestinationInteractive() error {
	// collect secrets list
	usClient := helpers.MustUpstreamClient()
	ussByKey := make(map[string]*v1.Upstream)
	var usKeys []string
	for _, ns := range helpers.MustGetNamespaces() {
		usList, err := usClient.List(ns, clients.ListOpts{})
		if err != nil {
			return err
		}
		for _, us := range usList {
			ref := us.Metadata.Ref()
			ussByKey[ref.Key()] = us
			usKeys = append(usKeys, ref.Key())
		}
	}
	if len(usKeys) == 0 {
		return errors.Errorf("no upstreams found. create an upstream first or enable " +
			"discovery.")
	}
	var usKey string
	if err := cliutil.ChooseFromList(
		"Choose the upstream to route to: ",
		&usKey,
		usKeys,
	); err != nil {
		return err
	}
	us, ok := ussByKey[usKey]
	if !ok {
		return errors.Errorf("internal error: upstream map not populated")
	}
}
