package surveyutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func surveyResources(resourceType, prompt, emptyKey string, list resources.ResourceList) (core.ResourceRef, error) {
	byKey := make(map[string]core.ResourceRef)
	var keys []string

	if emptyKey != "" {
		keys = append(keys, emptyKey)
	}

	for _, resource := range list {
		ref := resource.GetMetadata().Ref()
		byKey[ref.Key()] = ref
		keys = append(keys, ref.Key())
	}

	if len(keys) == 0 {
		return core.ResourceRef{}, errors.Errorf("no %v found. create one first.", resourceType)
	}

	var key string
	if err := cliutil.ChooseFromList(
		prompt,
		&key,
		keys,
	); err != nil {
		return core.ResourceRef{}, err
	}

	if key == emptyKey {
		return core.ResourceRef{}, nil
	}

	ref, ok := byKey[key]
	if !ok {
		return core.ResourceRef{}, errors.Errorf("internal error: %v map missing key %v", resourceType, key)
	}

	return ref, nil
}
