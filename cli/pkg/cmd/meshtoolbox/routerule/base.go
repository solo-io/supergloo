package routerule

import (
	"fmt"
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/solo-kit/pkg/errors"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	glooV1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
)

func EnsureMinimumRequiredParams(opts *options.Options) error {
	rrOpts := &(opts.Create).InputRoutingRule

	// all operations require a target mesh spec
	if err := nsutil.EnsureMesh(&rrOpts.TargetMesh, opts); err != nil {
		return err
	}

	// Validate source and destination upstreams
	if err := EnsureUpstreams(opts); err != nil {
		return err
	}

	// Validate matchers
	opts.MeshTool.RoutingRule.RequestMatchers = []*glooV1.Matcher{}
	if rrOpts.Matchers != nil {
		if err := ValidateMatchers(rrOpts.Matchers, opts.MeshTool.RoutingRule.RequestMatchers); err != nil {
			return err
		}

	}
	return nil
}

func EnsureUpstreams(opts *options.Options) error {
	rrOpts := &(opts.Create).InputRoutingRule
	if opts.Top.Static {
		upstreamClient, err := common.GetUpstreamClient()
		if err != nil {
			return err
		}
		var sources []*glooV1.Upstream
		if rrOpts.Sources != "" {
			sources, err = validateUpstreams(upstreamClient, rrOpts.Sources)
			if err != nil {
				return err
			}
			opts.MeshTool.RoutingRule.Sources = toResourceRefs(sources)
		}
		var destinations []*glooV1.Upstream
		if rrOpts.Destinations != "" {
			destinations, err = validateUpstreams(upstreamClient, rrOpts.Destinations)
			if err != nil {
				return err
			}
			opts.MeshTool.RoutingRule.Destinations = toResourceRefs(destinations)
		}
	} else {
		if err := nsutil.EnsureCommonResources("upstream", "Please select source upstream(s)", &(opts.MeshTool.RoutingRule).Sources, opts); err != nil {
			return err
		}
		if err := nsutil.EnsureCommonResources("upstream", "Please select destination upstream(s)", &(opts.MeshTool.RoutingRule).Destinations, opts); err != nil {
			return err
		}

	}
	return nil
}

func toResourceRefs(upstreams []*glooV1.Upstream) []*core.ResourceRef {
	if upstreams == nil {
		return nil
	}
	refs := make([]*core.ResourceRef, len(upstreams))
	for i, u := range upstreams {
		refs[i] = &core.ResourceRef{
			Name:      u.Metadata.Name,
			Namespace: u.Metadata.Namespace,
		}
	}
	return refs
}

func validateUpstreams(client *glooV1.UpstreamClient, upstreamOption string) ([]*glooV1.Upstream, error) {
	sources := strings.Split(upstreamOption, common.ListOptionSeparator)
	upstreams := make([]*glooV1.Upstream, len(sources))
	// TODO validate namespace?
	for i, u := range sources {
		parts := strings.Split(u, common.NamespacedResourceSeparator)
		if len(parts) != 2 {
			return nil, fmt.Errorf(common.InvalidOptionFormat, u, "create routingrule")
		}
		namespace, name := parts[0], parts[1]
		upstream, err := (*client).Read(namespace, name, clients.ReadOpts{})
		if err != nil {
			return nil, err
		}
		upstreams[i] = upstream
	}

	return upstreams, nil
}

func ValidateMatchers(matcherOption []string, targetMatchers []*glooV1.Matcher) error {
	result := make([]*glooV1.Matcher, len(matcherOption))

	// Each 'matcher' is one occurrence of the --matchers flag and will result in one glooV1.Matcher object
	// e.g. matcher = "prefix=/some/path,methods=get|post"
	for i, matcher := range matcherOption {
		m := &glooV1.Matcher{}

		// 'clauses' is an array of matcher clauses
		// e.g. clauses = { "prefix=/some/path" , "methods=get|post" }
		clauses := strings.Split(matcher, common.ListOptionSeparator)
		for _, clause := range clauses {

			// 'parts' are the two parts of the clause, separated by the "=" character
			// e.g. parts = { "prefix", "/some/path"}
			parts := strings.Split(clause, "=")
			if len(parts) != 2 {
				return fmt.Errorf(common.InvalidOptionFormat, clause, "create routingrule")
			}

			matcherType, value := parts[0], parts[1]
			switch matcherType {
			// TODO: make these (and other string literals around here) constants
			case "prefix":

				// We can have only one path specifier per matcher
				if m.PathSpecifier != nil {
					return errors.Errorf(common.InvalidOptionFormat, clause, "create routingrule")
				}
				m.PathSpecifier = &glooV1.Matcher_Prefix{Prefix: value}

			case "methods":

				// If the user specified more than one "methods" clause, return error. We could just merge the two
				// clauses, but this scenario is most likely an error we want the user to be aware of.
				if m.Methods != nil {
					return errors.Errorf(common.InvalidOptionFormat, clause, "create routingrule")
				}

				methods := strings.Split(value, common.SubListOptionSeparator)
				validMethods := strings.Split(common.ValidMatcherHttpMethods, common.SubListOptionSeparator)
				for _, method := range methods {
					if !common.Contains(validMethods, strings.ToUpper(method)) {
						return errors.Errorf(common.InvalidMatcherHttpMethod, method)
					}
				}
				m.Methods = methods

			default:
				return fmt.Errorf(common.InvalidOptionFormat, clause, "create routingrule")
			}
		}

		result[i] = m
	}

	targetMatchers = result
	return nil
}
