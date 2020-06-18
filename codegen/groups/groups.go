package groups

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	smhModule       = "github.com/solo-io/service-mesh-hub"
	v1alpha1Version = "v1alpha1"
	apiRoot         = "pkg/api"
)

var SMHGroups = []model.Group{
	makeGroup("core", v1alpha1Version, []resourceToGenerate{
		{kind: "Settings", noStatus: true},
	}),
	makeGroup("discovery", v1alpha1Version, []resourceToGenerate{
		{kind: "KubernetesCluster", noStatus: true}, // TODO(ilackarms): remove this kubernetes cluster and use skv2 multicluster
		{kind: "MeshService"},
		{kind: "MeshWorkload"},
		{kind: "Mesh"},
	}),
	makeGroup("networking", v1alpha1Version, []resourceToGenerate{
		{kind: "TrafficPolicy"},
		{kind: "AccessControlPolicy"},
		{kind: "VirtualMesh"},
	}),
}

var CSRGroups = []model.Group{
	makeGroup("security", v1alpha1Version, []resourceToGenerate{
		{kind: "VirtualMeshCertificateSigningRequest"},
	}),
}

type resourceToGenerate struct {
	kind     string
	noStatus bool // don't put a status on this resource
}

func makeGroup(groupPrefix, version string, resourcesToGenerate []resourceToGenerate) model.Group {
	var resources []model.Resource
	for _, resource := range resourcesToGenerate {
		res := model.Resource{
			Kind: resource.kind,
			Spec: model.Field{
				Type: model.Type{
					Name: resource.kind + "Spec",
				},
			},
		}
		if !resource.noStatus {
			res.Status = &model.Field{Type: model.Type{
				Name: resource.kind + "Status",
			}}
		}
		resources = append(resources, res)
	}

	return model.Group{
		GroupVersion: schema.GroupVersion{
			Group:   groupPrefix + "." + constants.ServiceMeshHubApiGroupSuffix,
			Version: version,
		},
		Module:           smhModule,
		Resources:        resources,
		RenderManifests:  true,
		RenderTypes:      true,
		RenderClients:    true,
		RenderController: true,
		RenderProtos:     true,
		MockgenDirective: true,
		CustomTemplates:  contrib.AllCustomTemplates,
		ApiRoot:          apiRoot,
	}
}
