package groups

/*
	DO NOT IMPORT THIS PACKAGE
	This package imports the "github.com/solo-io/skv2/contrib" package, which
	will panic when skv2 template files are not found in the executing environment.
*/

import (
	"github.com/solo-io/gloo-mesh/codegen/constants"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	glooMeshModule  = "github.com/solo-io/gloo-mesh"
	glooMeshApiRoot = "pkg/api"
)

var GlooMeshSettingsGroup = makeGroup("settings", "v1", []ResourceToGenerate{
	{Kind: "Settings", ShortNames: []string{"s"}},
	{Kind: "Dashboard", ShortNames: []string{"dash", "dashes"}},
})

var GlooMeshDiscoveryGroup = makeGroup("discovery", "v1", []ResourceToGenerate{
	{Kind: "Destination", ShortNames: []string{"dest", "dests"}},
	{Kind: "Workload", ShortNames: []string{"wkld", "wklds"}},
	{Kind: "Mesh"},
})

var GlooMeshNetworkingGroup = makeGroup("networking", "v1", []ResourceToGenerate{
	{Kind: "TrafficPolicy", ShortNames: []string{"tp", "tps"}},
	{Kind: "AccessPolicy", ShortNames: []string{"ap", "aps"}},
	{Kind: "VirtualMesh", ShortNames: []string{"vm", "vms"}},
})

var GlooMeshEnterpriseNetworkingGroup = makeGroup("networking.enterprise", "v1beta1", []ResourceToGenerate{
	{Kind: "WasmDeployment", ShortNames: []string{"wd", "wds"}},
	{Kind: "RateLimiterServerConfig", ShortNames: []string{"rlsc", "rlscs"}},
	{Kind: "VirtualDestination", ShortNames: []string{"vd", "vds"}},
	{Kind: "VirtualGateway", ShortNames: []string{"vg", "vgs"}},
	{Kind: "VirtualHost", ShortNames: []string{"vh", "vhs"}},
	{Kind: "RouteTable", ShortNames: []string{"rt", "rts"}},
	{Kind: "ServiceDependency", ShortNames: []string{"sd", "sds"}},
})

var GlooMeshEnterpriseObservabilityGroup = makeGroup("observability.enterprise", "v1", []ResourceToGenerate{
	{Kind: "AccessLogRecord", ShortNames: []string{"alr", "alrs"}},
})

var GlooMeshEnterpriseRbacGroup = makeGroup("rbac.enterprise", "v1", []ResourceToGenerate{
	{Kind: "Role", ShortNames: []string{"gmrole", "gmroles"}},
	{Kind: "RoleBinding", ShortNames: []string{"gmrolebinding", "gmrolebindings"}},
})

var GlooMeshGroups = []model.Group{
	GlooMeshEnterpriseNetworkingGroup,
	GlooMeshNetworkingGroup,
	GlooMeshSettingsGroup,
	GlooMeshDiscoveryGroup,
	GlooMeshEnterpriseObservabilityGroup,
	GlooMeshEnterpriseRbacGroup,
}

var CertAgentGroups = []model.Group{
	makeGroup("certificates", "v1", []ResourceToGenerate{
		{Kind: "IssuedCertificate"},
		{Kind: "CertificateRequest"},
		{Kind: "PodBounceDirective"},
	}),
}

var XdsAgentGroup = makeGroup("xds.agent.enterprise", "v1beta1", []ResourceToGenerate{
	{Kind: "XdsConfig"},
})

var AllGeneratedGroups = append(
	append(
		GlooMeshGroups,
		CertAgentGroups...,
	),
	XdsAgentGroup,
)

type ResourceToGenerate struct {
	Kind       string
	ShortNames []string
	NoStatus   bool // don't put a status on this resource
}

func makeGroup(groupPrefix, version string, resourcesToGenerate []ResourceToGenerate) model.Group {
	return MakeGroup(glooMeshModule, glooMeshApiRoot, groupPrefix, version, resourcesToGenerate)
}

// exported for use in enterprise repo
func MakeGroup(module, apiRoot, groupPrefix, version string, resourcesToGenerate []ResourceToGenerate) model.Group {
	var resources []model.Resource
	for _, resource := range resourcesToGenerate {
		res := model.Resource{
			Kind: resource.Kind,
			Spec: model.Field{
				Type: model.Type{
					Name: resource.Kind + "Spec",
				},
			},
			ShortNames: resource.ShortNames,
		}
		if !resource.NoStatus {
			res.Status = &model.Field{Type: model.Type{
				Name: resource.Kind + "Status",
			}}
		}
		resources = append(resources, res)
	}

	return model.Group{
		GroupVersion: schema.GroupVersion{
			Group:   groupPrefix + "." + constants.GlooMeshApiGroupSuffix,
			Version: version,
		},
		Module:                  module,
		Resources:               resources,
		RenderManifests:         true,
		RenderValidationSchemas: true,
		RenderTypes:             true,
		RenderClients:           true,
		RenderController:        true,
		MockgenDirective:        true,
		CustomTemplates:         contrib.AllGroupCustomTemplates,
		ApiRoot:                 apiRoot,
	}
}
