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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RoutingRuleCmd(opts *options.RoutingRule) *cobra.Command {
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
			fmt.Printf("Created routing rule [%v] in namespace [%v]\n", args[0], opts.Namespace)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Mesh, "mesh", "", "", "The mesh that will be the target for this rule")
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default",
		"The namespace for this routing rule. Defaults to \"default\"")
	flags.StringVarP(&opts.Sources, "sources", "", "", "Sources for this rule. Each entry "+
		"consists of an upstream namespace and and upstream name, separated by a colon.")
	flags.StringVarP(&opts.Destinations, "destinations", "", "", "Destinations for this rule. Same format as for 'sources'")
	flags.StringVarP(&opts.Matchers, "matchers", "", "", "Matchers for this rule")
	flags.BoolVarP(&opts.OverrideExisting, "override", "", false, "If set to \"true\", "+
		"the command will override any existing routing rule that matches the given namespace and name")

	// The only required option is the target mesh
	cmd.MarkFlagRequired("mesh")

	return cmd
}

func createRoutingRule(routeName string, opts *options.RoutingRule) error {

	// Ensure that the given mesh exists
	meshClient, err := common.GetMeshClient()
	if err != nil {
		return err
	}
	mesh, err := (*meshClient).Read(constants.SuperglooNamespace, opts.Mesh, clients.ReadOpts{})
	if err != nil {
		return err
	}

	// Validate namespace
	if opts.Namespace != "" && opts.Namespace != "default" {
		kube, err := common.GetKubernetesClient()
		if err != nil {
			return err
		}
		_, err = kube.CoreV1().Namespaces().Get(opts.Namespace, v1.GetOptions{IncludeUninitialized: false})
		if err != nil {
			return err
		}
	}

	// Validate source and destination upstreams
	upstreamClient, err := common.GetUpstreamClient()
	if err != nil {
		return err
	}
	var sources []*glooV1.Upstream
	if opts.Sources != "" {
		sources, err = validateUpstreams(upstreamClient, opts.Sources)
		if err != nil {
			return err
		}
	}
	var destinations []*glooV1.Upstream
	if opts.Destinations != "" {
		sources, err = validateUpstreams(upstreamClient, opts.Destinations)
		if err != nil {
			return err
		}
	}

	// Validate matchers
	var matchers []*glooV1.Matcher
	if opts.Matchers != "" {
		matchers, err = validateMatchers(opts.Matchers)
	}
	if err != nil {
		return err
	}

	routingRule := &superglooV1.RoutingRule{
		Metadata: core.Metadata{
			Name:      routeName,
			Namespace: opts.Namespace,
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
	_, err = (*rrClient).Write(routingRule, clients.WriteOpts{OverwriteExisting: opts.OverrideExisting})
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
