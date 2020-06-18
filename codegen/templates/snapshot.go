package templates

import (
	"io/ioutil"
	"log"
	"text/template"

	"github.com/solo-io/skv2/codegen/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const outputSnapshotTemplatePath = "codegen/templates/output_snapshot.gotmpl"

var OutputSnapshotTemplateContents = func() string {
	b, err := ioutil.ReadFile(outputSnapshotTemplatePath)
	if err != nil {
		log.Fatalf("failed to read file %v", err)
	}
	return string(b)
}()

// make the custom template funcs
func MakeSnapshotFuncs(groups []model.Group) template.FuncMap {
	return template.FuncMap{
		"snapshot_groups": func() []model.Group { return groups },
	}
}

func SelectResources(groups []model.Group, resourcesToSelect map[schema.GroupVersion][]string) []model.Group {
	var selectedResources []model.Group
	for _, group := range groups {
		resources := resourcesToSelect[group.GroupVersion]
		if len(resources) == 0 {
			continue
		}
		filteredGroup := group
		filteredGroup.Resources = nil

		isResourceSelected := func(kind string) bool {
			for _, resource := range resources {
				if resource == kind {
					return true
				}
			}
			return false
		}

		for _, resource := range group.Resources {
			if !isResourceSelected(resource.Kind) {
				continue
			}
			filteredGroup.Resources = append(filteredGroup.Resources, resource)
		}

		selectedResources = append(selectedResources, filteredGroup)
	}

	return selectedResources
}
