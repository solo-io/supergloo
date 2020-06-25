package templates

import (
	"strings"
	"text/template"

	"github.com/solo-io/skv2/codegen/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type importedGroup struct {
	model.Group
	GoModule string // the module where the group is defined, if it differs from the group module itself. e.g. for external type imports such as k8s.io/api
}

// make the custom template funcs
func MakeSnapshotFuncs(importedGroups []importedGroup) template.FuncMap {
	var groups []model.Group
	groupImports := map[schema.GroupVersion]importedGroup{}

	for _, grp := range importedGroups {
		groups = append(groups, grp.Group)
		groupImports[grp.GroupVersion] = grp
	}

	return template.FuncMap{
		"imported_groups": func() []model.Group { return groups },
		"client_import_path": func(group model.Group) string {
			grp, ok := groupImports[group.GroupVersion]
			if !ok {
				panic("group not found " + grp.String())
			}
			return clientImportPath(grp)
		},
		"set_import_path": func(group model.Group) string {
			grp, ok := groupImports[group.GroupVersion]
			if !ok {
				panic("group not found " + grp.String())
			}
			return clientImportPath(grp) + "/sets"
		},
		"controller_import_path": func(group model.Group) string {
			grp, ok := groupImports[group.GroupVersion]
			if !ok {
				panic("group not found " + grp.String())
			}
			return clientImportPath(grp) + "/controller"
		},
	}
}

// gets the go package for an imported group's clients
func clientImportPath(grp importedGroup) string {

	grp.ApiRoot = strings.Trim(grp.ApiRoot, "/")

	module := grp.GoModule
	if module == "" {
		// import should be our local module, which comes from the imported group
		module = grp.Group.Module
	}

	s := strings.ReplaceAll(
		strings.Join([]string{
			module,
			grp.ApiRoot,
			grp.Group.Group,
			grp.Version,
		}, "/"),
		"//", "/",
	)

	return s
}

// pass empty string if clients live in the same go module as the type definitions
func SelectResources(clientModule string, groups []model.Group, resourcesToSelect map[schema.GroupVersion][]string) []importedGroup {
	var selectedResources []importedGroup
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

		selectedResources = append(selectedResources, importedGroup{
			Group:    filteredGroup,
			GoModule: clientModule,
		})
	}

	return selectedResources
}
