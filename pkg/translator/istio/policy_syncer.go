package istio

import (
	"context"
	"fmt"
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type PolicySyncer struct {
}

func (s *PolicySyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {

}

func c() {
	rcfg := &v1alpha1.RbacConfig{
		Mode:            v1alpha1.RbacConfig_Mode,
		EnforcementMode: v1alpha1.EnforcementMode_ENFORCED,
	}
}

func toIstio(p *v1.Policy) ([]*v1alpha1.ServiceRole, []*v1alpha1.ServiceRoleBinding) {
	var roles []*v1alpha1.ServiceRole
	var bindings []*v1alpha1.ServiceRoleBinding

	rulesByDest := map[core.ResourceRef][]*v1.Rule{}
	for _, rule := range p.Rules {
		rulesByDest[rule.Destination] = append(rulesByDest[rule.Destination], rule)
	}
	// sort for idempotency
	for _, rule := range p.Rules {
		dests := rulesByDest[rule.Destination]
		sort.Slice(dests, func(i, j int) bool {
			return dests[i].Source.String() > dests[j].Source.String()
		})
	}

	for dest, rules := range rulesByDest {
		ns := rule.Destination.Metadata.Namespace
		// create an istio service role and binding:
		name := "access-" + rule.Destination.Metadata.Namespace + "-" + rule.Destination.Metadata.Name
		// create service role:
		sr := &v1alpha1.ServiceRole{
			Metadata: core.Metadata{
				Name:      name,
				Namespace: ns,
			},
			Rules: []*v1alpha1.AccessRule{
				{
					Services: []string{
						svcname(rule.Destination),
					},
				},
			},
		}
		var subjects []*v1alpha1.Subject
		for _, rule := range rules {
			subjects = append(subjects, *v1alpha1.Subject{
				Properties: map[string]string{
					"source.principal", principalame(rule.Source),
				},
			})
		}
		name = "bind-" + rule.Destination.Metadata.Namespace + "-" + rule.Destination.Metadata.Name
		srb := &v1alpha1.ServiceRoleBinding{
			Metadata: core.Metadata{
				Name:      name,
				Namespace: ns,
			},
			Subjects: subjects,
			RoleRef: &v1alpha1.RoleRef{
				Name: sr.Metadata.Name,
			},
		}
		roles = append(roles, sr)
		bindings = append(bindings, srb)
	}
	return roles, bindings
}

func svcname(s core.ResourceRef) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace)
}
func principalame(s core.ResourceRef) string {
	return fmt.Sprintf("cluster.local/ns/%s/sa/%s", s.Namespace, s.Name)
}
