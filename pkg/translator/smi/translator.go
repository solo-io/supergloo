package smi

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type Translator interface {
	// translates a snapshot into a set of istio configs for each mesh
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error)
}

// A container for the entire set of config for a single istio mesh
type MeshConfig struct {
	RoutingConfig  *RoutingConfig
	SecurityConfig *SecurityConfig
}

func (c *MeshConfig) Sort() {
	c.RoutingConfig.Sort()
	c.SecurityConfig.Sort()
}

// first create all destination rules for all subsets of each upstream
// then we need to apply the ISTIO_MUTUAL policy depending on
// whether mtls is enabled

type translator struct {
}

func NewTranslator() Translator {
	return &translator{}
}

/*
Translate a snapshot into a set of MeshConfigs for each mesh
Currently only active istio mesh is expected.
*/
func (t *translator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*MeshConfig, reporter.ResourceErrors, error) {

	meshes := snapshot.Meshes
	meshGroups := snapshot.Meshgroups
	upstreams := snapshot.Upstreams
	services := snapshot.Services
	pods := snapshot.Pods
	routingRules := snapshot.Routingrules
	securityRules := snapshot.Securityrules

	resourceErrs := make(reporter.ResourceErrors)

	utils.ValidateMeshGroups(meshes, meshGroups, resourceErrs)

	routingRulesByMesh := utils.GroupRulesByMesh(routingRules, securityRules, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]*MeshConfig)

	for _, mesh := range meshes {
		istio := mesh.GetIstio()
		if istio == nil || !mesh.GetSmiEnabled() {
			continue
		}
		resourceErrs.Accept(mesh)

		writeNamespace := istio.InstallationNamespace
		rules := routingRulesByMesh[mesh]
		in := inputMeshConfig{
			writeNamespace: writeNamespace,
			mesh:           mesh,
			rules:          rules,
		}
		meshConfig, err := t.translateMesh(in, upstreams, services, pods, resourceErrs)
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
	input inputMeshConfig,
	upstreams gloov1.UpstreamList,
	services kubernetes.ServiceList,
	pods kubernetes.PodList,
	resourceErrs reporter.ResourceErrors) (*MeshConfig, error) {

	routingConfig := createRoutingConfig(
		input.rules.Routing,
		upstreams,
		services,
		resourceErrs,
	)

	securityConfig := createSecurityConfig(
		input.rules.Security,
		upstreams,
		pods,
		services,
		resourceErrs,
	)

	meshConfig := &MeshConfig{
		RoutingConfig:  &routingConfig,
		SecurityConfig: &securityConfig,
	}
	meshConfig.Sort()

	return meshConfig, nil
}
