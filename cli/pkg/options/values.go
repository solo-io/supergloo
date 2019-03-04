package options

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/vektah/gqlgen/neelance/errors"
)

type ResourceRefsValue []core.ResourceRef

func (v *ResourceRefsValue) String() string {
	if v == nil {
		return "<nil>"
	}
	var strs []string
	for _, r := range *v {
		strs = append(strs, fmt.Sprintf("%v.%v", r.Namespace, r.Name))
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

func (v *ResourceRefsValue) Set(s string) error {
	split := strings.SplitN(s, ".", 2)
	if len(split) != 2 {
		return errors.Errorf("%s invalid: refs must be specified in the format <NAMESPACE>.<NAME>", s)
	}
	*v = append(*v, core.ResourceRef{Namespace: split[0], Name: split[1]})
	return nil
}

func (v *ResourceRefsValue) Type() string {
	return "ResourceRefsValue"
}

type MapStringStringValue map[string]string

func (v *MapStringStringValue) String() string {
	if v == nil {
		return "<nil>"
	}
	var strs []string
	for k, val := range *v {
		strs = append(strs, fmt.Sprintf("%v.%v", k, val))
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

func (v *MapStringStringValue) Set(s string) error {
	split := strings.SplitN(s, "=", 2)
	if len(split) != 2 {
		return errors.Errorf("%s invalid: map entries must be specified in the format KEY=VALUE", s)
	}
	m := *v
	m[split[0]] = split[1]
	*v = m
	return nil
}

func (v *MapStringStringValue) Type() string {
	return "MapStringStringValue"
}

type RequestMatchersValue []RequestMatcher

func (v *RequestMatchersValue) String() string {
	if v == nil {
		return "<nil>"
	}
	var strs []string
	for _, r := range *v {
		strs = append(strs, fmt.Sprintf("%#v", r))
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

func (v *RequestMatchersValue) Set(s string) error {
	var match RequestMatcher
	err := json.Unmarshal([]byte(s), &match)
	if err != nil {
		return err
	}
	*v = append(*v, match)
	return nil
}

func (v *RequestMatchersValue) Type() string {
	return "RequestMatchersValue"
}
