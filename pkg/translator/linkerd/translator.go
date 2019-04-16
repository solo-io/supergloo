package linkerd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	linkerdv1alpha1 "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/linkerd/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type Translator interface {
	// translates a snapshot into a set of istio configs for each mesh
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error)
}

// A container for the entire set of config for a single istio mesh
type MeshConfig struct {
	ServiceProfiles v1.ServiceProfileList
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
Currently only active istio mesh is expected.
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
	pods := snapshot.Pods.List()
	routingRules := snapshot.Routingrules.List()
	securityRules := snapshot.Securityrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)

	utils.ValidateMeshGroups(meshes, meshGroups, resourceErrs)

	routingRulesByMesh := utils.GroupRulesByMesh(routingRules, securityRules, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]*MeshConfig)

	params := plugins.Params{
		Ctx:       ctx,
		Upstreams: upstreams,
	}

	tlsSecrets := snapshot.Tlssecrets.List()

	for _, mesh := range meshes {
		istio, ok := mesh.MeshType.(*v1.Mesh_Istio)
		if !ok {
			// we only want istio meshes
			continue
		}
		writeNamespace := istio.Istio.InstallationNamespace
		rules := routingRulesByMesh[mesh]
		in := inputMeshConfig{
			writeNamespace: writeNamespace,
			mesh:           mesh,
			rules:          rules,
		}
		meshConfig, err := t.translateMesh(params, in, upstreams, tlsSecrets, pods, resourceErrs)
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

// produces a complete istio config
func (t *translator) translateMesh(
	params plugins.Params,
	input inputMeshConfig,
	upstreams gloov1.UpstreamList,
	tlsSecrets v1.TlsSecretList,
	pods v1.PodList,
	resourceErrs reporter.ResourceErrors) (*MeshConfig, error) {

	rules := input.rules

	// get all destinations that have routing rules
	destinationsWithRoutingRules := make(map[*gloov1.Upstream]v1.RoutingRuleList)
	for _, rr := range rules.Routing {
		dests, err := utils.UpstreamsForSelector(rr.DestinationSelector, upstreams)
		if err != nil {
			return nil, err
		}
		for _, dest := range dests {
			destinationsWithRoutingRules[dest] = append(destinationsWithRoutingRules[dest], rr)
		}
	}

	// create a service profile for each upstream with routing rules
	var serviceProfiles v1.ServiceProfileList
	for dest, rules := range destinationsWithRoutingRules {
		sp := initServiceProfile(dest.GetMetadata().Ref())
		sp.SetMetadata(core.Metadata{
			Namespace: dest.GetMetadata().Namespace,
			Name:      dest.GetMetadata().Name,
		})

		// process plugins for each rule
		for _, rr := range rules {
			if rr.Spec == nil {
				resourceErrs.AddError(rr, errors.Errorf("spec cannot be empty"))
				continue
			}

			routes := convertMatchers(rr.RequestMatchers)
			sp.Spec.Routes = append(sp.Spec.Routes, routes...)

			for _, plug := range t.plugins {
				switch plug := plug.(type) {
				case plugins.ServiceProfilePlugin:
					if err := plug.ProcessServiceProfile(params, *rr.Spec, &sp.Spec); err != nil {
						resourceErrs.AddError(rr, errors.Wrapf(err, "processing routes"))
						continue
					}
				case plugins.RoutingPlugin:
					if err := plug.ProcessRoutes(params, *rr.Spec, routes); err != nil {
						resourceErrs.AddError(rr, errors.Wrapf(err, "processing routes"))
						continue
					}
				}
			}

		}

		serviceProfiles = append(serviceProfiles, sp)
	}

	meshConfig := &MeshConfig{
		ServiceProfiles: serviceProfiles,
	}
	meshConfig.Sort()

	return meshConfig, nil
}

func initServiceProfile(forUpstream core.ResourceRef) *v1.ServiceProfile {
	sp := &v1.ServiceProfile{}
	sp.SetMetadata(core.Metadata{
		Namespace: forUpstream.Namespace,
		Name:      forUpstream.Name,
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
