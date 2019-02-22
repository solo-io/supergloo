package surveyutils

import (
	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/gloo/projects/gloo/cli/pkg/helpers"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func SurveyMetadata(meta *core.Metadata) error {
	if err := cliutil.GetStringInput("name for the resource: ", &meta.Name); err != nil {
		return err
	}
	if err := cliutil.ChooseFromList("namespace for the resource: ", &meta.Namespace, helpers.MustGetNamespaces()); err != nil {
		return err
	}
	return nil
}
