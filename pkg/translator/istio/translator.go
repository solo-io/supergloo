package istio

import (
	"context"
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/hashutils"
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
	DesinationRules v1alpha3.DestinationRuleList
	VirtualServices v1alpha3.VirtualServiceList
	MeshPolicy      *v1alpha1.MeshPolicy // meshpolicy is a singleton
}

func (c *MeshConfig) Sort() {
	sort.SliceStable(c.DesinationRules, func(i, j int) bool {
		return c.DesinationRules[i].Metadata.Less(c.DesinationRules[j].Metadata)
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

// we create a routing rule for each unique matcher
type rulesByMatcher struct {
	rules map[uint64]v1.RoutingRuleList
}

func newRulesByMatcher(rules v1.RoutingRuleList) rulesByMatcher {
	rbm := make(map[uint64]v1.RoutingRuleList)
	for _, rule := range rules {
		hash := hashutils.HashAll(
			rule.SourceSelector,
			rule.RequestMatchers,
		)
		rbm[hash] = append(rbm[hash], rule)
	}

	return rulesByMatcher{rules: rbm}
}

func (rbm rulesByMatcher) sort() []v1.RoutingRuleList {
	var (
		hashes       []uint64
		rulesForHash []v1.RoutingRuleList
	)
	for hash, rules := range rbm.rules {
		hashes = append(hashes, hash)
		rulesForHash = append(rulesForHash, rules)
	}
	sort.SliceStable(rulesForHash, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})
	return rulesForHash
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
	// list of route rules which apply to this mesh
	rules v1.RoutingRuleList
}

type rulesByMesh map[*v1.Mesh]v1.RoutingRuleList

func splitRulesByMesh(rules v1.RoutingRuleList, meshes v1.MeshList, meshGroups v1.MeshGroupList, resourceErrs reporter.ResourceErrors) rulesByMesh {
	rulesByMesh := make(rulesByMesh)
	for _, rule := range rules {
		targetMesh := rule.TargetMesh
		if targetMesh == nil {
			resourceErrs.AddError(rule, errors.Errorf("target mesh cannot be nil"))
			continue
		}
		mesh, err := meshes.Find(targetMesh.Strings())
		if err == nil {
			rulesByMesh[mesh] = append(rulesByMesh[mesh], rule)
			continue
		}
		meshGroup, err := meshGroups.Find(targetMesh.Strings())
		if err != nil {
			resourceErrs.AddError(rule, errors.Errorf("no target mesh or mesh group found for %v", targetMesh))
			continue
		}
		for _, ref := range meshGroup.Meshes {
			if ref == nil {
				resourceErrs.AddError(meshGroup, errors.Errorf("ref cannot be nil"))
				resourceErrs.AddError(rule, errors.Errorf("referenced invalid MeshGroup %v", meshGroup.Metadata.Ref()))
				continue
			}
			mesh, err := meshes.Find(ref.Strings())
			if err != nil {
				resourceErrs.AddError(meshGroup, err)
				resourceErrs.AddError(rule, errors.Errorf("referenced invalid MeshGroup %v", meshGroup.Metadata.Ref()))
				continue
			}
			rulesByMesh[mesh] = append(rulesByMesh[mesh], rule)
		}
	}
	return rulesByMesh
}

func (t *translator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error) {
	meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	routingRules := snapshot.Routingrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)
	resourceErrs.Accept(upstreams.AsInputResources()...)

	routingRulesByMesh := splitRulesByMesh(routingRules, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]*MeshConfig)

	params := plugins.Params{
		Ctx: ctx,
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
		meshConfig, err := t.translateMesh(params, in, upstreams, resourceErrs)
		if err != nil {
			return nil, nil, err
		}
		perMeshConfig[mesh] = meshConfig
	}

	return perMeshConfig, resourceErrs, nil
}

// produces a complete istio config
func (t *translator) translateMesh(params plugins.Params, input inputMeshConfig, upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors) (*MeshConfig, error) {
	ctx := params.Ctx
	mtlsEnabled := input.mesh.MtlsConfig != nil && input.mesh.MtlsConfig.MtlsEnabled
	routingRules := input.rules

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

		dr := t.makeDestinatioRuleForHost(ctx,
			params,
			input.writeNamespace,
			destinationHost,
			labelSets,
			mtlsEnabled,
			resourceErrs,
		)
		destinationRules = append(destinationRules, dr)

		vs := t.makeVirtualServiceForHost(ctx,
			params,
			input.writeNamespace,
			destinationHost,
			destinationPortAndLabelSets,
			routingRules,
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

	meshConfig := &MeshConfig{
		VirtualServices: virtualServices,
		DesinationRules: destinationRules,
		MeshPolicy:      meshPolicy,
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
