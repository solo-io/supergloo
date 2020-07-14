package codegen

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	V1alpha1Version = "v1alpha1"
	SmhModule       = "github.com/solo-io/service-mesh-hub"
	ApiRoot         = "pkg/api"
	SmhGroups       = []model.Group{
		MakeGroup("core", V1alpha1Version, []ResourceToGenerate{
			{Kind: "Settings", NoStatus: true},
		}),
		MakeGroup("discovery", V1alpha1Version, []ResourceToGenerate{
			{Kind: "KubernetesCluster", NoStatus: true}, // TODO(ilackarms): remove this kubernetes cluster and use skv2 multicluster
			{Kind: "MeshService"},
			{Kind: "MeshWorkload"},
			{Kind: "Mesh"},
		}),
		MakeGroup("networking", V1alpha1Version, []ResourceToGenerate{
			{Kind: "TrafficPolicy"},
			{Kind: "AccessControlPolicy"},
			{Kind: "VirtualMesh"},
			{Kind: "FailoverService"},
		}),
	}
)

type ResourceToGenerate struct {
	Kind     string
	NoStatus bool // don't put a status on this resource
}

func MakeGroup(groupPrefix, version string, resourcesToGenerate []ResourceToGenerate) model.Group {
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
			Group:   groupPrefix + "." + constants.ServiceMeshHubApiGroupSuffix,
			Version: version,
		},
		Module:           SmhModule,
		Resources:        resources,
		RenderManifests:  true,
		RenderTypes:      true,
		RenderClients:    true,
		RenderController: true,
		MockgenDirective: true,
		CustomTemplates:  contrib.AllGroupCustomTemplates,
		ApiRoot:          ApiRoot,
	}
}
