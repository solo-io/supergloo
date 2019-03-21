package surveyutils

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func SurveyMetadata(resourceName string, meta *core.Metadata) error {
	if err := cliutil.GetStringInput(fmt.Sprintf("name for the %v: ", resourceName), &meta.Name); err != nil {
		return err
	}
	if err := cliutil.ChooseFromList(fmt.Sprintf("namespace for the %v: ", resourceName), &meta.Namespace, clients.MustGetNamespaces()); err != nil {
		return err
	}
	return nil
}
