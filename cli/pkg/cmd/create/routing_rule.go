package create

import (
	"fmt"
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/solo-kit/pkg/errors"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	glooV1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func RoutingRuleCmd(opts *options.Options) *cobra.Command {
	rrOpts := opts.Create.RoutingRule
	cmd := &cobra.Command{
		Use:   "routingrule",
		Short: `Create a route rule with the given name`,
		Long:  `Create a route rule with the given name`,
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			if err := createRoutingRule(args[0], opts); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Created routing rule [%v] in namespace [%v]\n", args[0], rrOpts.Namespace)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&rrOpts.Mesh, "mesh", "", "", "The mesh that will be the target for this rule")
	flags.StringVarP(&rrOpts.Namespace, "namespace", "n", "default",
		"The namespace for this routing rule. Defaults to \"default\"")
	flags.StringVarP(&rrOpts.Sources, "sources", "", "", "Sources for this rule. Each entry "+
		"consists of an upstream namespace and and upstream name, separated by a colon.")
	flags.StringVarP(&rrOpts.Destinations, "destinations", "", "", "Destinations for this rule. Same format as for 'sources'")
	flags.StringVarP(&rrOpts.Matchers, "matchers", "", "", "Matchers for this rule")
	flags.BoolVarP(&rrOpts.OverrideExisting, "override", "", false, "If set to \"true\", "+
		"the command will override any existing routing rule that matches the given namespace and name")

	// The only required option is the target mesh
	cmd.MarkFlagRequired("mesh")

	return cmd
}

func createRoutingRule(routeName string, opts *options.Options) error {
	rrOpts := opts.Create.RoutingRule

	// Ensure that the given mesh exists
	meshClient, err := common.GetMeshClient()
	if err != nil {
		return err
	}
	mesh, err := (*meshClient).Read(constants.SuperglooNamespace, rrOpts.Mesh, clients.ReadOpts{})
	if err != nil {
		return err
	}

	// Validate namespace
	if rrOpts.Namespace != "" && rrOpts.Namespace != "default" {
		if common.Contains(opts.Cache.Namespaces, rrOpts.Namespace) {
			return fmt.Errorf("Namespace %v does not exist.\n", rrOpts.Namespace)
		}
	}

	// Validate source and destination upstreams
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
	}
	var destinations []*glooV1.Upstream
	if rrOpts.Destinations != "" {
		sources, err = validateUpstreams(upstreamClient, rrOpts.Destinations)
		if err != nil {
			return err
		}
	}

	// Validate matchers
	var matchers []*glooV1.Matcher
	if rrOpts.Matchers != "" {
		matchers, err = validateMatchers(rrOpts.Matchers)
	}
	if err != nil {
		return err
	}

	routingRule := &superglooV1.RoutingRule{
		Metadata: core.Metadata{
			Name:      routeName,
			Namespace: rrOpts.Namespace,
		},
		TargetMesh: &core.ResourceRef{
			Name:      mesh.Metadata.Name,
			Namespace: mesh.Metadata.Namespace,
		},
		Sources:         toResourceRefs(sources),
		Destinations:    toResourceRefs(destinations),
		RequestMatchers: matchers,
	}

	rrClient, err := common.GetRoutingRuleClient()
	if err != nil {
		return err
	}
	_, err = (*rrClient).Write(routingRule, clients.WriteOpts{OverwriteExisting: rrOpts.OverrideExisting})
	return err
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

func validateMatchers(matcherOption string) ([]*glooV1.Matcher, error) {
	matchers := strings.Split(matcherOption, common.ListOptionSeparator)
	result := make([]*glooV1.Matcher, len(matchers))

	for i, m := range matchers {

		parts := strings.Split(m, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf(common.InvalidOptionFormat, m, "create routingrule")
		}

		matcherType, value := parts[0], parts[1]
		switch matcherType {

		// TODO: make these (and other string literals around here) constants
		case "prefix":
			// TODO(marco) validate if valid URL path?
			result[i] = &glooV1.Matcher{PathSpecifier: &glooV1.Matcher_Prefix{Prefix: value}}

		case "methods":
			methods := strings.Split(value, "|")
			validMethods := strings.Split(common.ValidMatcherHttpMethods, "|")
			for _, method := range methods {
				if !common.Contains(validMethods, strings.ToUpper(method)) {
					return nil, errors.Errorf(common.InvalidMatcherHttpMethod, method)
				}
			}
			result[i] = &glooV1.Matcher{Methods: methods}

		default:
			return nil, fmt.Errorf(common.InvalidOptionFormat, m, "create routingrule")
		}
	}
	return result, nil
}
