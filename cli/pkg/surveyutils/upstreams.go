package surveyutils

import (
	"context"

	"github.com/solo-io/gloo/pkg/cliutil"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
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
		us, err := selectUpstreamInteractive("add an upstream (choose <done> to finish): ", upstreams)
		if err != nil {
			return nil, err
		}
		// the user chose
		if us == nil {
			return selected, nil
		}
		selected = append(selected, us.Metadata.Ref())
	}
}

const skipSelector = "<done>"

func selectUpstreamInteractive(prompt string, upstreams v1.UpstreamList) (*v1.Upstream, error) {
	ussByKey := make(map[string]*v1.Upstream)
	usKeys := []string{skipSelector}
	for _, us := range upstreams {
		ref := us.Metadata.Ref()
		ussByKey[ref.Key()] = us
		usKeys = append(usKeys, ref.Key())
	}

	if len(usKeys) == 1 {
		return nil, errors.Errorf("no upstreams found. create an upstream first or enable " +
			"discovery.")
	}

	var usKey string
	if err := cliutil.ChooseFromList(
		prompt,
		&usKey,
		usKeys,
	); err != nil {
		return nil, err
	}

	// user skipped
	if usKey == skipSelector {
		return nil, nil
	}

	us, ok := ussByKey[usKey]
	if !ok {
		return nil, errors.Errorf("internal error: upstream map missing key %v", usKey)
	}

	return us, nil
}
