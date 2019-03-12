package istio

import (
	"context"
	"sort"

	"github.com/solo-io/go-utils/contextutils"

	customkube "github.com/solo-io/supergloo/pkg/api/external/kubernetes/core/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type Translator interface {
	// translates a snapshot into a set of istio configs for each mesh
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error)
}

// A container for the entire set of config for a single istio mesh
type MeshConfig struct {
	// mtls
	MeshPolicy *v1alpha1.MeshPolicy // meshpolicy is a singleton

	// routing
	DestinationRules v1alpha3.DestinationRuleList
	VirtualServices  v1alpha3.VirtualServiceList

	// rbac
	SecurityConfig
}

func (c *MeshConfig) Sort() {
	sort.SliceStable(c.DestinationRules, func(i, j int) bool {
		return c.DestinationRules[i].Metadata.Less(c.DestinationRules[j].Metadata)
	})
	sort.SliceStable(c.VirtualServices, func(i, j int) bool {
		return c.VirtualServices[i].Metadata.Less(c.VirtualServices[j].Metadata)
	})
}

// first create all desintation rules for all subsets of each upstream
// then we need to apply the ISTIO_MUTUAL policy depending on
// whether mtls is enabled

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

	// TODO (ilackarms): when we support installing Gloo
	// ensure that we handle race condition with upstream
	// reporting.
	resourceErrs.Accept(upstreams.AsInputResources()...)

	validateMeshGroups(meshes, meshGroups, resourceErrs)

	routingRulesByMesh := splitRulesByMesh(ctx, routingRules, securityRules, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]*MeshConfig)

	params := plugins.Params{
		Ctx:       ctx,
		Upstreams: upstreams,
	}

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
		meshConfig, err := t.translateMesh(params, in, upstreams, pods, resourceErrs)
		if err != nil {
			resourceErrs.AddError(mesh, errors.Wrapf(err, "translating mesh config"))
			contextutils.LoggerFrom(ctx).Errorf("translating for mesh %v failed: %v", mesh.Metadata.Ref(), err)
			continue
		}
		perMeshConfig[mesh] = meshConfig
	}

	return perMeshConfig, resourceErrs, nil
}

type labelsPortTuple struct {
	labels map[string]string
	port   uint32
}

type inputMeshConfig struct {
	// where crds should be written. this is normally the mesh installation namespace
	writeNamespace string
	// the mesh we're configuring
	mesh *v1.Mesh
	// list of rules which apply to this mesh
	rules ruleSet
}

type meshRule interface {
	resources.InputResource
	GetTargetMesh() *core.ResourceRef
}

type rulesByMesh map[*v1.Mesh]ruleSet

func (rbm rulesByMesh) addRule(ctx context.Context, rule meshRule, meshes v1.MeshList, meshGroups v1.MeshGroupList) error {
	var appendRule func(*v1.Mesh)

	switch r := rule.(type) {
	case *v1.RoutingRule:
		appendRule = func(mesh *v1.Mesh) {
			rule := rbm[mesh]
			rule.routing = append(rule.routing, r)
			rbm[mesh] = rule
		}
	case *v1.SecurityRule:
		appendRule = func(mesh *v1.Mesh) {
			rule := rbm[mesh]
			rule.security = append(rule.security, r)
			rbm[mesh] = rule
		}
	default:
		return errors.Errorf("internal error: cannot append rule type %v", rule)
	}

	targetMesh := rule.GetTargetMesh()
	if targetMesh == nil {
		return errors.Errorf("target mesh cannot be nil")
	}
	mesh, err := meshes.Find(targetMesh.Strings())
	if err == nil {
		appendRule(mesh)
		return nil
	}
	meshGroup, err := meshGroups.Find(targetMesh.Strings())
	if err != nil {
		return errors.Errorf("no target mesh or mesh group found for %v", targetMesh)
	}
	for _, ref := range meshGroup.Meshes {
		if ref == nil {
			return errors.Errorf("referenced invalid MeshGroup %v", meshGroup.Metadata.Ref())
		}
		mesh, err := meshes.Find(ref.Strings())
		if err != nil {
			return errors.Errorf("referenced invalid MeshGroup %v", meshGroup.Metadata.Ref())
		}
		appendRule(mesh)
	}

	return nil
}

type ruleSet struct {
	routing  v1.RoutingRuleList
	security v1.SecurityRuleList
}

func splitRulesByMesh(ctx context.Context, routingRules v1.RoutingRuleList, securityRules v1.SecurityRuleList, meshes v1.MeshList, meshGroups v1.MeshGroupList, resourceErrs reporter.ResourceErrors) rulesByMesh {
	rulesByMesh := make(rulesByMesh)

	for _, rule := range routingRules {
		if err := rulesByMesh.addRule(ctx, rule, meshes, meshGroups); err != nil {
			resourceErrs.AddError(rule, err)
			continue
		}
	}
	for _, rule := range securityRules {
		if err := rulesByMesh.addRule(ctx, rule, meshes, meshGroups); err != nil {
			resourceErrs.AddError(rule, err)
			continue
		}
	}
	return rulesByMesh
}

// TODO (ilackarms) move to the top-level translator
func validateMeshGroups(meshes v1.MeshList, meshGroups v1.MeshGroupList, resourceErrs reporter.ResourceErrors) {
	for _, mg := range meshGroups {
		for _, ref := range mg.Meshes {
			if ref == nil {
				resourceErrs.AddError(mg, errors.Errorf("ref cannot be nil"))
				continue
			}
			if _, err := meshes.Find(ref.Strings()); err != nil {
				resourceErrs.AddError(mg, err)
				continue
			}
		}
	}
}

// produces a complete istio config
func (t *translator) translateMesh(
	params plugins.Params,
	input inputMeshConfig,
	upstreams gloov1.UpstreamList,
	pods customkube.PodList,
	resourceErrs reporter.ResourceErrors) (*MeshConfig, error) {
	ctx := params.Ctx
	mtlsEnabled := input.mesh.MtlsConfig != nil && input.mesh.MtlsConfig.MtlsEnabled
	rules := input.rules

	destinationHostsPortsAndLabels, err := labelsAndPortsByHost(upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "internal error: getting ports and labels from upstreams")
	}
	var destinationRules v1alpha3.DestinationRuleList
	var virtualServices v1alpha3.VirtualServiceList
	for destinationHost, destinationPortAndLabelSets := range destinationHostsPortsAndLabels {
		var labelSets []map[string]string
		for _, set := range destinationPortAndLabelSets {
			labelSets = append(labelSets, set.labels)
		}

		dr := makeDestinationRule(ctx,
			input.writeNamespace,
			destinationHost,
			labelSets,
			mtlsEnabled,
		)
		destinationRules = append(destinationRules, dr)

		vs := t.makeVirtualServiceForHost(ctx,
			params,
			input.writeNamespace,
			destinationHost,
			destinationPortAndLabelSets,
			rules.routing,
			upstreams,
			resourceErrs,
		)

		virtualServices = append(virtualServices, vs)
	}

	var meshPolicy *v1alpha1.MeshPolicy
	if mtlsEnabled {
		meshPolicy = &v1alpha1.MeshPolicy{
			Metadata: core.Metadata{
				// the required name for istio MeshPolicy
				// https://istio.io/docs/tasks/security/authn-policy/#globally-enabling-istio-mutual-tls
				Name: "default",
			},
			Peers: []*v1alpha1.PeerAuthenticationMethod{{
				Params: &v1alpha1.PeerAuthenticationMethod_Mtls{Mtls: &v1alpha1.MutualTls{
					Mode: v1alpha1.MutualTls_STRICT,
				}},
			}},
		}
	}

	securityConfig := createSecurityConfig(
		input.writeNamespace,
		input.rules.security,
		upstreams,
		pods,
		resourceErrs,
	)

	meshConfig := &MeshConfig{
		VirtualServices:  virtualServices,
		DestinationRules: destinationRules,
		MeshPolicy:       meshPolicy,
		SecurityConfig:   securityConfig,
	}
	meshConfig.Sort()

	return meshConfig, nil
}

func labelsAndPortsByHost(upstreams gloov1.UpstreamList) (map[string][]labelsPortTuple, error) {
	labelsByHost := make(map[string][]labelsPortTuple)
	for _, us := range upstreams {
		labels := utils.GetLabelsForUpstream(us)
		host, err := utils.GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		port, err := utils.GetPortForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		labelsByHost[host] = append(labelsByHost[host], labelsPortTuple{labels: labels, port: port})
	}
	return labelsByHost, nil
}
