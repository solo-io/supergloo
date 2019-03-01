package surveyutils

import (
	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/vektah/gqlgen/neelance/errors"
)

func SurveyUninstall(opts *options.Options) error {
	installs, err := helpers.MustInstallClient().List("", clients.ListOpts{Ctx: opts.Ctx})
	if err != nil {
		return err
	}
	var activeInstalls v1.InstallList
	byName := make(map[string]*v1.Install)
	for _, in := range installs {
		if !in.Disabled {
			activeInstalls = append(activeInstalls, in)
			byName[in.Metadata.Namespace+"."+in.Metadata.Name] = in
		}
	}
	if len(activeInstalls) == 0 {
		return errors.Errorf("no active installs found")
	}
	var namespaceDotName string
	if err := cliutil.ChooseFromList("which install to uninstall? ", &namespaceDotName, activeInstalls.NamespacesDotNames()); err != nil {
		return err
	}

	opts.Uninstall.Metadata = byName[namespaceDotName].Metadata

	return nil
}
