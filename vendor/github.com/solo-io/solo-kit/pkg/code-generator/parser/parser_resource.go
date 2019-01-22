package parser

import (
	"sort"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/iancoleman/strcase"
	"github.com/ilackarms/protokit"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

const (
	// solo-kit types
	// required fields
	metadataTypeName = ".core.solo.io.Metadata"
	statusTypeName   = ".core.solo.io.Status"

	// magic comments
	// Deprecated, use Message Option (core.solo.io.resource).short_name
	shortNameDeclaration = "@solo-kit:resource.short_name="
	// Deprecated, use Message Option (core.solo.io.resource).plural_name
	pluralNameDeclaration = "@solo-kit:resource.plural_name="
	// Deprecated, use projectConfig.ResourceGroups
	resourceGroupsDeclaration = "@solo-kit:resource.resource_groups="
)

// add some data we need to the regular proto message
type ProtoMessageWrapper struct {
	GoPackage string
	Message   *protokit.Descriptor
}

// note (ilackarms): this function supports the deprecated method of using magic comments to declare resource groups.
// this will be removed in a future release of solo kit
func resourceGroupsFromMessages(messages []ProtoMessageWrapper) map[string][]model.ResourceConfig {
	resourceGroupsCfg := make(map[string][]model.ResourceConfig)
	for _, msg := range messages {
		comments := strings.Split(msg.Message.GetComments().Leading, "\n")
		// optional flags
		joinedResourceGroups, _ := getCommentValue(comments, resourceGroupsDeclaration)
		resourceGroups := strings.Split(joinedResourceGroups, ",")
		for _, rgName := range resourceGroups {
			if rgName == "" {
				continue
			}
			resourceGroupsCfg[rgName] = append(resourceGroupsCfg[rgName], model.ResourceConfig{
				MessageName:    msg.Message.GetName(),
				MessagePackage: msg.Message.GetPackage(),
			})
		}
	}
	return resourceGroupsCfg
}

func getResource(resources []*model.Resource, cfg model.ResourceConfig) (*model.Resource, error) {
	for _, res := range resources {
		if res.Name == cfg.MessageName && res.ProtoPackage == cfg.MessagePackage {
			return res, nil
		}
	}
	return nil, errors.Errorf("getting resource: message %v not found", cfg)
}

func getResources(project *model.Project, messages []ProtoMessageWrapper) ([]*model.Resource, []*model.ResourceGroup, error) {
	// legacy behavior (deprecated): if resource groups are not specified, search through protos for
	// resourceGroupsDeclaration
	if len(project.ProjectConfig.ResourceGroups) == 0 {
		project.ProjectConfig.ResourceGroups = resourceGroupsFromMessages(messages)
	}
	var (
		resources []*model.Resource
	)
	for _, msg := range messages {
		resource, err := describeResource(msg)
		if err != nil {
			return nil, nil, err
		}
		if resource == nil {
			// not a solo-kit resource, ignore
			continue
		}
		resource.Project = project
		resources = append(resources, resource)
	}

	var (
		resourceGroups []*model.ResourceGroup
	)

	for groupName, resourcesCfg := range project.ProjectConfig.ResourceGroups {
		var resourcesForGroup []*model.Resource
		for _, resourceCfg := range resourcesCfg {
			resource, err := getResource(resources, resourceCfg)
			if err != nil {
				return nil, nil, err
			}
			if resource.ProtoPackage != project.ProtoPackage {
				importPrefix := strings.Replace(resource.ProtoPackage, ".", "_", -1) + "."
				resource.ImportPrefix = importPrefix
			}
			resourcesForGroup = append(resourcesForGroup, resource)
		}

		log.Printf("creating resource group: %v", groupName)
		rg := &model.ResourceGroup{
			Name:      groupName,
			GoName:    goName(groupName),
			Project:   project,
			Resources: resourcesForGroup,
		}
		for _, res := range resourcesForGroup {
			res.ResourceGroups = append(res.ResourceGroups, rg)
		}

		imports := make(map[string]string)
		for _, res := range rg.Resources {
			// only generate files for the resources in our group, otherwise we import
			if res.ProtoPackage != rg.Project.ProtoPackage {
				// add import
				imports[strings.TrimSuffix(res.ImportPrefix, ".")] = res.GoPackage
			}
		}
		var sortedImports []string
		for k := range imports {
			sortedImports = append(sortedImports, k)
		}
		sort.Strings(sortedImports)
		for _, imp := range sortedImports {
			rg.Imports += imp + " \"" + imports[imp] + "\"\n\t"
		}

		resourceGroups = append(resourceGroups, rg)
	}

	// sort for stability
	for _, res := range resources {
		sort.SliceStable(res.ResourceGroups, func(i, j int) bool {
			return res.ResourceGroups[i].Name < res.ResourceGroups[j].Name
		})
	}
	sort.SliceStable(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	sort.SliceStable(resourceGroups, func(i, j int) bool {
		return resourceGroups[i].Name < resourceGroups[j].Name
	})
	return resources, resourceGroups, nil
}

func describeResource(messageWrapper ProtoMessageWrapper) (*model.Resource, error) {
	msg := messageWrapper.Message
	// not a solo kit resource, or you messed up!
	if !hasField(msg, "metadata", metadataTypeName) {
		return nil, nil
	}

	comments := strings.Split(msg.GetComments().Leading, "\n")

	name := msg.GetName()
	var (
		shortName, pluralName string
		clusterScoped         bool
	)
	resourceOpts, err := proto.GetExtension(msg.Options, core.E_Resource)
	if err != nil {
		log.Warnf("failed to get solo-kit message options for resource %v: %v", msg.GetName(), err)
		log.Warnf("use of magic comments is deprecated, use Message Option (core.solo.io.resource)")
		// required flags
		sn, ok := getCommentValue(comments, shortNameDeclaration)
		if !ok {
			return nil, errors.Errorf("must provide %s", shortNameDeclaration)
		}
		shortName = sn
		pn, ok := getCommentValue(comments, pluralNameDeclaration)
		if !ok {
			return nil, errors.Errorf("must provide %s", pluralNameDeclaration)
		}
		pluralName = pn
	} else {
		res, ok := resourceOpts.(*core.Resource)
		if !ok {
			return nil, errors.Errorf("internal error: message options were not type *core.Resource: %+v", resourceOpts)
		}
		shortName = res.ShortName
		pluralName = res.PluralName
		clusterScoped = res.ClusterScoped
	}

	// always make it upper camel
	pluralName = strcase.ToCamel(pluralName)

	hasStatus := hasField(msg, "status", statusTypeName)

	fields := collectFields(msg)
	oneofs := collectOneofs(msg)

	return &model.Resource{
		Name:          name,
		ProtoPackage:  msg.GetPackage(),
		GoPackage:     messageWrapper.GoPackage,
		ShortName:     shortName,
		PluralName:    pluralName,
		HasStatus:     hasStatus,
		Fields:        fields,
		Oneofs:        oneofs,
		ClusterScoped: clusterScoped,
		Filename:      msg.GetFile().GetName(),
		Original:      msg,
	}, nil
}
