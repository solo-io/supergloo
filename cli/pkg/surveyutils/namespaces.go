package surveyutils

import (
	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
)

const doneOption = "<done>"

func SurveyNamespaces() ([]string, error) {
	allNs := append([]string{doneOption}, helpers.MustGetNamespaces()...)
	var selected []string

	for {
		var ns string
		if err := cliutil.ChooseFromList("add a namespace (choose <done> to finish): ", &ns, allNs); err != nil {
			return nil, err
		}

		// the user chose done
		if ns == doneOption {
			return selected, nil
		}
		selected = append(selected, ns)
	}
}
