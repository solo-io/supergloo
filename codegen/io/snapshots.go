package io

import (
	"sort"
	"strings"

	"github.com/gertd/go-pluralize"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// a Snapshot is a group of individual resources from one or more GroupVersions
type Snapshot map[schema.GroupVersion][]string

type OutputSnapshot struct {
	// Snapshot to be used as an output snapshot
	Snapshot Snapshot
	// Name to be used for placing this builder in a subdirectory
	Name string
}

// get the rbac policies needed to watch the snapshot
func (s Snapshot) RbacPoliciesWatch() []rbacv1.PolicyRule {
	return s.rbacPolicies(
		[]string{"get", "list", "watch"},
		"",
	)
}

// get the rbac policies needed to write the snapshot
func (s Snapshot) RbacPoliciesWrite() []rbacv1.PolicyRule {
	return s.rbacPolicies(
		[]string{"*"},
		"",
	)
}

// get the rbac policies needed to update the snapshot statuses
func (s Snapshot) RbacPoliciesUpdateStatus() []rbacv1.PolicyRule {
	return s.rbacPolicies(
		[]string{"get", "update"},
		"status",
	)
}

func (s Snapshot) rbacPolicies(
	verbs []string,
	subresource string,
) []rbacv1.PolicyRule {
	var policies []rbacv1.PolicyRule
	for groupVersion, kinds := range s {
		var resources []string
		for _, kind := range kinds {
			resource := strings.ToLower(pluralize.NewClient().Plural(kind))
			if subresource != "" {
				resource += "/" + subresource
			}
			resources = append(resources, resource)
		}
		policies = append(policies, rbacv1.PolicyRule{
			Verbs:     verbs,
			APIGroups: []string{groupVersion.Group},
			Resources: resources,
		})
	}
	sort.SliceStable(policies, func(i, j int) bool {
		return policies[i].APIGroups[0] < policies[j].APIGroups[0]
	})
	return policies
}
