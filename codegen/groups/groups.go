package groups

import (
	"github.com/solo-io/gloo-mesh/codegen/constants"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	glooMeshModule  = "github.com/solo-io/gloo-mesh"
	v1alpha2Version = "v1alpha2"
	glooMeshApiRoot = "pkg/api"
)

var GlooMeshGroups = []model.Group{
	makeGroup("settings", v1alpha2Version, []ResourceToGenerate{
		{Kind: "Settings"},
	}),
	makeGroup("discovery", v1alpha2Version, []ResourceToGenerate{
		{Kind: "TrafficTarget"},
		{Kind: "Workload"},
		{Kind: "Mesh"},
	}),
	makeGroup("networking", v1alpha2Version, []ResourceToGenerate{
		{Kind: "TrafficPolicy"},
		{Kind: "AccessPolicy"},
		{Kind: "VirtualMesh"},
		{Kind: "FailoverService"},
	}),
}

var CertAgentGroups = []model.Group{
	makeGroup("certificates", v1alpha2Version, []ResourceToGenerate{
		{Kind: "IssuedCertificate"},
		{Kind: "CertificateRequest"},
		{Kind: "PodBounceDirective", NoStatus: true},
	}),
}

type ResourceToGenerate struct {
	Kind     string
	NoStatus bool // don't put a status on this resource
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
