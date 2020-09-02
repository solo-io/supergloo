package appmesh

import (
	"text/template"

	"github.com/solo-io/service-mesh-hub/codegen/utils"
	"github.com/solo-io/skv2/codegen/model"
)

func MakeTemplates(appmeshResources []string) model.CustomTemplates {
	funcs := template.FuncMap{
		"appmesh_resources": func() []string {
			return appmeshResources
		},
		"comparison_options": func(resource string) string {
			switch resource {
			case "VirtualNode":
				// special comparison option for VirtualNodes
				return "equality.CompareOptionForVirtualNodeSpec"
			default:
				return "cmpopts.EquateEmpty"
			}
		},
	}

	return model.CustomTemplates{
		Templates: map[string]string{
			"pkg/api/external/appmesh/appmesh_client.go":   utils.MustReadFile("appmesh_client.gotmpl"),
			"pkg/api/external/appmesh/appmesh_snapshot.go": utils.MustReadFile("appmesh_snapshot.gotmpl"),
		},
		MockgenDirective: true,
		Funcs:            funcs,
	}
}
