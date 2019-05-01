package linkerd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	linkerdv1alpha1 "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"
	"github.com/linkerd/linkerd2/pkg/profiles"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	linkerdv1 "github.com/solo-io/supergloo/pkg/api/external/linkerd/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/linkerd/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
	"github.com/solo-io/supergloo/pkg/util"
	"sigs.k8s.io/yaml"
)

type Translator interface {
	// translates a snapshot into a set of linkerd configs for each mesh
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error)
}

// A container for the entire set of config for a single linkerd mesh
type MeshConfig struct {
	ServiceProfiles linkerdv1.ServiceProfileList
}

func (c *MeshConfig) Sort() {
	sort.SliceStable(c.ServiceProfiles, func(i, j int) bool {
		return c.ServiceProfiles[i].GetMetadata().Less(c.ServiceProfiles[j].GetMetadata())
	})
}

type translator struct {
	plugins []plugins.Plugin
}

func NewTranslator(plugins []plugins.Plugin) Translator {
	return &translator{plugins: plugins}
}

/*
Translate a snapshot into a set of MeshConfigs for each mesh
Currently only active linkerd mesh is expected.
*/
func (t *translator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error) {

	// initialize plugins
	initParams := plugins.InitParams{
		Ctx: ctx,
	}
	for _, plug := range t.plugins {
		if err := plug.Init(initParams); err != nil {
			return nil, nil, err
		}
	}

	meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	routingRules := snapshot.Routingrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)

	utils.ValidateMeshGroups(meshes, meshGroups, resourceErrs)

	routingRulesByMesh := utils.GroupRulesByMesh(routingRules, nil, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]*MeshConfig)

	params := plugins.Params{
		Ctx:       ctx,
		Upstreams: upstreams,
	}

	for _, mesh := range meshes {
		linkerd := mesh.GetLinkerd()
		if linkerd == nil {
			// we only want linkerd meshes
			continue
		}

		writeNamespace := util.GetMeshInstallatioNamespace(mesh)
		rules := routingRulesByMesh[mesh]
		in := inputMeshConfig{
			writeNamespace: writeNamespace,
			mesh:           mesh,
			rules:          rules,
		}
		meshConfig, err := t.translateMesh(params, in, upstreams, resourceErrs)
		if err != nil {
			resourceErrs.AddError(mesh, errors.Wrapf(err, "translating mesh config"))
			contextutils.LoggerFrom(ctx).Errorf("translating for mesh %v failed: %v", mesh.Metadata.Ref(), err)
			continue
		}
		perMeshConfig[mesh] = meshConfig
	}

	return perMeshConfig, resourceErrs, nil
}

type inputMeshConfig struct {
	// where crds should be written. this is normally the mesh installation namespace
	writeNamespace string
	// the mesh we're configuring
	mesh *v1.Mesh
	// list of rules which apply to this mesh
	rules utils.RuleSet
}

type hostWithRules struct {
	host      string
	namespace string
	rules     v1.RoutingRuleList
}

type hostsWithRules []*hostWithRules

func (hosts *hostsWithRules) addRule(rr *v1.RoutingRule, dest *gloov1.Upstream) error {
	host, err := utils.GetHostForUpstream(dest)
	if err != nil {
		return err
	}
	for _, existingHost := range *hosts {
		if existingHost.host == host {
			for _, existingRule := range existingHost.rules {
				if existingRule == rr {
					return nil
				}
			}
			existingHost.rules = append(existingHost.rules, rr)
			return nil
		}
	}
	*hosts = append(*hosts, &hostWithRules{
		host:      host,
		namespace: utils.GetNamespaceForUpstream(dest),
		rules:     v1.RoutingRuleList{rr},
	})
	return nil
}

// produces a complete linkerd config
func (t *translator) translateMesh(
	params plugins.Params,
	input inputMeshConfig,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors) (*MeshConfig, error) {

	rules := input.rules

	// get all destinations that have routing rules
	hosts := &hostsWithRules{}
	for _, rr := range rules.Routing {
		dests, err := utils.UpstreamsForSelector(rr.DestinationSelector, upstreams)
		if err != nil {
			return nil, err
		}
		for _, dest := range dests {
			if err := hosts.addRule(rr, dest); err != nil {
				return nil, err
			}
		}
	}

	// create a service profile for each upstream with routing rules
	var serviceProfiles linkerdv1.ServiceProfileList
	for _, hostRules := range *hosts {
		sp := initServiceProfile(hostRules.host, hostRules.namespace)

		// process plugins for each rule
		for _, rr := range hostRules.rules {
			if rr.Spec == nil {
				resourceErrs.AddError(rr, errors.Errorf("spec cannot be empty"))
				continue
			}

			routes := convertMatchers(rr.RequestMatchers)
			sp.Spec.Routes = append(sp.Spec.Routes, routes...)

			for _, plug := range t.plugins {
				if plug, ok := plug.(plugins.ServiceProfilePlugin); ok {
					if err := plug.ProcessServiceProfile(params, *rr.Spec, &sp.Spec); err != nil {
						resourceErrs.AddError(rr, errors.Wrapf(err, "processing routes"))
					}
				}
				if plug, ok := plug.(plugins.RoutingPlugin); ok {
					if err := plug.ProcessRoutes(params, *rr.Spec, routes); err != nil {
						resourceErrs.AddError(rr, errors.Wrapf(err, "processing routes"))
					}
				}
			}
		}

		// only create service profiles with some kind of config
		if sp.Spec.RetryBudget == nil && len(sp.Spec.Routes) == 0 {
			continue
		}

		if err := validateServiceProfile(sp); err != nil {
			resourceErrs.AddError(input.mesh, errors.Wrapf(err, "internal error: produced invalid service profile"))
			continue
		}

		serviceProfiles = append(serviceProfiles, sp)
	}

	meshConfig := &MeshConfig{
		ServiceProfiles: serviceProfiles,
	}
	meshConfig.Sort()

	return meshConfig, nil
}

func initServiceProfile(hostname, namespace string) *linkerdv1.ServiceProfile {
	sp := &linkerdv1.ServiceProfile{}
	sp.SetMetadata(core.Metadata{
		Name:      hostname,
		Namespace: namespace,
	})
	return sp
}

func convertMatchers(matchers []*gloov1.Matcher) []*linkerdv1alpha1.RouteSpec {
	var ldRoutes []*linkerdv1alpha1.RouteSpec
	for _, match := range matchers {
		ldRoutes = append(ldRoutes, convertMatcher(match))
	}
	return ldRoutes
}

func convertMatcher(match *gloov1.Matcher) *linkerdv1alpha1.RouteSpec {
	var pathRegex string
	if match.PathSpecifier != nil {
		switch path := match.PathSpecifier.(type) {
		case *gloov1.Matcher_Exact:
			pathRegex = path.Exact
		case *gloov1.Matcher_Regex:
			pathRegex = path.Regex
		case *gloov1.Matcher_Prefix:
			pathRegex = path.Prefix + ".*"
		}
	}
	methods := match.Methods

	// if the user specified > 1 method, create an Any matcher
	if len(methods) > 1 {
		var matchConditions []*linkerdv1alpha1.RequestMatch
		for _, method := range methods {
			matchConditions = append(matchConditions, &linkerdv1alpha1.RequestMatch{
				PathRegex: pathRegex,
				Method:    method,
			})
		}
		return &linkerdv1alpha1.RouteSpec{
			Name:      fmt.Sprintf("%v %v", strings.Join(methods, ","), pathRegex),
			Condition: &linkerdv1alpha1.RequestMatch{Any: matchConditions},
		}
	}

	if len(methods) == 0 {
		methods = []string{""} // match all methods
	}

	method := methods[0]
	return &linkerdv1alpha1.RouteSpec{
		Name: fmt.Sprintf("%v %v", method, pathRegex),
		Condition: &linkerdv1alpha1.RequestMatch{
			PathRegex: pathRegex,
			Method:    method,
		},
	}
}

func validateServiceProfile(sp *linkerdv1.ServiceProfile) error {
	raw, err := yaml.Marshal(sp)
	if err != nil {
		return err
	}
	return profiles.Validate(raw)
}
