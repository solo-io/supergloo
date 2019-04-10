package appmesh

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

const (
	PodVirtualNodeEnvName = "APPMESH_VIRTUAL_NODE_NAME"
	PodPortsEnvName       = "APPMESH_APP_PORTS"
)

type Translator interface {
	Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]AwsAppMeshConfiguration, reporter.ResourceErrors, error)
}

type appMeshTranslator struct{}

func NewAppMeshTranslator() Translator {
	return &appMeshTranslator{}
}

func (t *appMeshTranslator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]AwsAppMeshConfiguration, reporter.ResourceErrors, error) {
	meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	pods := snapshot.Pods.List()
	routingRules := snapshot.Routingrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)

	utils.ValidateMeshGroups(meshes, meshGroups, resourceErrs)

	// We currently don't handle security rules for App Mesh
	routingRulesByMesh := utils.GroupRulesByMesh(routingRules, nil, meshes, meshGroups, resourceErrs)

	perMeshConfig := make(map[*v1.Mesh]AwsAppMeshConfiguration)
	for _, mesh := range meshes {
		if mesh.GetAwsAppMesh() == nil {
			// Skip if not of type AppMesh
			continue
		}

		appMeshConfig, err := t.translateMesh(ctx, mesh, routingRulesByMesh[mesh].Routing, upstreams, pods, resourceErrs)
		if err != nil {
			resourceErrs.AddError(mesh, errors.Wrapf(err, "translating mesh config"))
			contextutils.LoggerFrom(ctx).Errorf("translating for mesh %v failed: %v", mesh.Metadata.Ref(), err)
			continue
		}
		perMeshConfig[mesh] = appMeshConfig
	}

	return perMeshConfig, resourceErrs, nil

}

func (t *appMeshTranslator) translateMesh(
	ctx context.Context,
	mesh *v1.Mesh,
	routingRules v1.RoutingRuleList,
	upstreams gloov1.UpstreamList,
	pods v1.PodList,
	resourceErrs reporter.ResourceErrors) (AwsAppMeshConfiguration, error) {

	config, err := NewAwsAppMeshConfiguration(mesh, pods, upstreams)
	if err != nil {
		return nil, err
	}

	if err := config.ProcessRoutingRules(routingRules); err != nil {
		return nil, err
	}

	if err := config.AllowAll(); err != nil {
		return nil, err
	}

	return config, err
}
