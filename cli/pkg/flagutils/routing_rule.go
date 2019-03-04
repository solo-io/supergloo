package flagutils

import (
	"encoding/json"
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddCreateRoutingRuleFlags(set *pflag.FlagSet, in *options.CreateRoutingRule) {
	addSelectorFlags("source", "originating from", set, &in.SourceSelector)
	addSelectorFlags("dest", "sent to", set, &in.DestinationSelector)
	addMatcherFlags(set, in)
}

func addSelectorFlags(prefix, direction string, set *pflag.FlagSet, in *options.Selector) {
	set.Var(&in.SelectedUpstreams,
		prefix+"-"+"upstreams",
		fmt.Sprintf("apply this rule to requests %v these upstreams. format must be <NAMESPACE>.<NAME>.", direction))

	set.Var(&in.SelectedLabels,
		prefix+"-"+"labels",
		fmt.Sprintf("apply this rule to requests %v pods with these labels. format must be KEY=VALUE", direction))

	set.StringSliceVar(&in.SelectedNamespaces,
		prefix+"-"+"namespaces",
		[]string{"default"},
		fmt.Sprintf("apply this rule to requests %v pods in these namespaces", direction))
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
