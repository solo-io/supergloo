package utils

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type meshRule interface {
	resources.InputResource
	GetTargetMesh() *core.ResourceRef
}

type RuleSet struct {
	Routing  v1.RoutingRuleList
	Security v1.SecurityRuleList
}

type RulesByMesh map[*v1.Mesh]RuleSet

func GroupRulesByMesh(
	routingRules v1.RoutingRuleList,
	securityRules v1.SecurityRuleList,
	meshes v1.MeshList,
	meshGroups v1.MeshGroupList,
	resourceErrs reporter.ResourceErrors) RulesByMesh {
	rulesByMesh := make(RulesByMesh)

	for _, rule := range routingRules {
		if err := rulesByMesh.addRule(rule, meshes, meshGroups); err != nil {
			resourceErrs.AddError(rule, err)
			continue
		}
	}
	for _, rule := range securityRules {
		if err := rulesByMesh.addRule(rule, meshes, meshGroups); err != nil {
			resourceErrs.AddError(rule, err)
			continue
		}
	}
	return rulesByMesh
}

func (rbm RulesByMesh) addRule(rule meshRule, meshes v1.MeshList, meshGroups v1.MeshGroupList) error {
	var appendRule func(*v1.Mesh)

	switch r := rule.(type) {
	case *v1.RoutingRule:
		appendRule = func(mesh *v1.Mesh) {
			rule := rbm[mesh]
			rule.Routing = append(rule.Routing, r)
			rbm[mesh] = rule
		}
	case *v1.SecurityRule:
		appendRule = func(mesh *v1.Mesh) {
			rule := rbm[mesh]
			rule.Security = append(rule.Security, r)
			rbm[mesh] = rule
		}
	default:
		return errors.Errorf("internal error: cannot append rule type %v", rule)
	}

	targetMesh := rule.GetTargetMesh()
	if targetMesh == nil {
		return errors.Errorf("target mesh cannot be nil")
	}
	mesh, err := meshes.Find(targetMesh.Strings())
	if err == nil {
		appendRule(mesh)
		return nil
	}
	meshGroup, err := meshGroups.Find(targetMesh.Strings())
	if err != nil {
		return errors.Errorf("no target mesh or mesh group found for %v", targetMesh)
	}
	for _, ref := range meshGroup.Meshes {
		if ref == nil {
			return errors.Errorf("referenced invalid MeshGroup %v", meshGroup.Metadata.Ref())
		}
		mesh, err := meshes.Find(ref.Strings())
		if err != nil {
			return errors.Errorf("referenced invalid MeshGroup %v", meshGroup.Metadata.Ref())
		}
		appendRule(mesh)
	}

	return nil
}
