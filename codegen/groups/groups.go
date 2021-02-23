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

var GlooMeshSettingsGroup = makeGroup("settings", "v1alpha2", []ResourceToGenerate{
	{Kind: "Settings"},
})

var GlooMeshDiscoveryGroup = makeGroup("discovery", "v1alpha2", []ResourceToGenerate{
	{Kind: "Destination"},
	{Kind: "Workload"},
	{Kind: "Mesh"},
})

var GlooMeshNetworkingGroup = makeGroup("networking", "v1alpha2", []ResourceToGenerate{
	{Kind: "TrafficPolicy"},
	{Kind: "AccessPolicy"},
	{Kind: "VirtualMesh"},
})

var GlooMeshEnterpriseNetworkingGroup = makeGroup("networking.enterprise", "v1alpha1", []ResourceToGenerate{
	{Kind: "WasmDeployment"},
	{Kind: "VirtualDestination"},
})

var GlooMeshEnterpriseObservabilityGroup = makeGroup("observability.enterprise", "v1alpha1", []ResourceToGenerate{
	{Kind: "AccessLogRecord"},
})

var GlooMeshEnterpriseRbacGroup = makeGroup("rbac.enterprise", "v1alpha1", []ResourceToGenerate{
	{Kind: "Role", ShortNames: []string{"gmrole", "gmroles"}},
	{Kind: "RoleBinding", ShortNames: []string{"gmrolebinding", "gmrolebindings"}},
})

var GlooMeshGroups = []model.Group{
	GlooMeshSettingsGroup,
	GlooMeshDiscoveryGroup,
	GlooMeshNetworkingGroup,
	GlooMeshEnterpriseNetworkingGroup,
	GlooMeshEnterpriseObservabilityGroup,
	GlooMeshEnterpriseRbacGroup,
}

var CertAgentGroups = []model.Group{
	makeGroup("certificates", "v1alpha2", []ResourceToGenerate{
		{Kind: "IssuedCertificate"},
		{Kind: "CertificateRequest"},
		{Kind: "PodBounceDirective"},
	}),
}

var XdsAgentGroup = makeGroup("xds.agent.enterprise", "v1alpha1", []ResourceToGenerate{
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
