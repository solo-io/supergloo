package surveyutils

import (
	"context"
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

const (
	SelectorType_Labels    = "Label Selector"
	SelectorType_Namespace = "Namespace Selector"
	SelectorType_Upstream  = "Upstream Selector"
)

func surveySelector(ctx context.Context, selectorName string, in *options.Selector) error {
	var enabled bool
	if err := cliutil.GetBoolInput(fmt.Sprintf("create a %v selector for this rule? ", selectorName), &enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	var selectorType string
	if err := cliutil.ChooseFromList("what kind of selector would you like to create? ", &selectorType, []string{
		SelectorType_Labels,
		SelectorType_Upstream,
		SelectorType_Namespace,
	}); err != nil {
		return err
	}
	switch selectorType {
	case SelectorType_Labels:
		fmt.Println("add key-value pairs to the selector:")
		m := map[string]string(in.SelectedLabels)
		if err := SurveyMapStringString(&m); err != nil {
			return err
		}
		in.SelectedLabels = options.MapStringStringValue(m)
	case SelectorType_Upstream:
		fmt.Println("choose upstreams for this selector")
		ups, err := SurveyUpstreams(ctx)
		if err != nil {
			return err
		}
		in.SelectedUpstreams = ups
	case SelectorType_Namespace:
		fmt.Println("choose namespaces for this selector")
		nss, err := SurveyNamespaces()
		if err != nil {
			return err
		}
		in.SelectedNamespaces = nss
	default:
		return errors.Errorf("%v is not a known selector type", selectorType)
	}

	return nil
}

func surveyMesh(ctx context.Context, mesh *options.ResourceRefValue) error {
	meshes, err := clients.MustMeshClient().List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return err
	}

	meshesByKey := make(map[string]core.ResourceRef)
	meshKeys := []string{}
	for _, mesh := range meshes {
		ref := mesh.Metadata.Ref()
		meshesByKey[ref.Key()] = ref
		meshKeys = append(meshKeys, ref.Key())
	}

	if len(meshKeys) == 0 {
		return errors.Errorf("no meshes found. register or install a mesh first.")
	}

	var key string
	if err := cliutil.ChooseFromList(
		"select a target mesh to which to apply this rule",
		&key,
		meshKeys,
	); err != nil {
		return err
	}

	ref, ok := meshesByKey[key]
	if !ok {
		return errors.Errorf("internal error: upstream map missing key %v", key)
	}

	*mesh = options.ResourceRefValue(ref)

	return nil
}
