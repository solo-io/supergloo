package flagutils

import (
	"encoding/json"

	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddCreateRoutingRuleFlags(set *pflag.FlagSet, in *options.CreateRoutingRule) {
	addSelectorFlags("source", "originating from", set, &in.SourceSelector)
	addSelectorFlags("dest", "sent to", set, &in.DestinationSelector)
	addMatcherFlags(set, in)
	addTargetMeshFlags(set, &in.TargetMesh)
}

var exampleMatcher = func() string {
	m := options.RequestMatcher{
		PathPrefix: "/users",
		Methods:    []string{"GET"},
		HeaderMatcher: map[string]string{
			"x-custom-header": "bar",
		},
	}
	b, _ := json.Marshal(m)
	return string(b)
}()

func addMatcherFlags(set *pflag.FlagSet, in *options.CreateRoutingRule) {
	set.Var(&in.RequestMatchers, "request-matcher", "json-formatted string which can be parsed as a "+
		"RequestMatcher type, e.g. "+exampleMatcher)
}
