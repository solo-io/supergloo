package istio

import (
	"fmt"

	"github.com/solo-io/go-utils/stringutils"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	"github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	customkube "github.com/solo-io/supergloo/pkg/api/external/kubernetes/core/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type SecurityConfig struct {
	RbacConfig          *v1alpha1.RbacConfig // singleton
	ServiceRoles        v1alpha1.ServiceRoleList
	ServiceRoleBindings v1alpha1.ServiceRoleBindingList
}

func createSecurityConfig(writeNamespace string,
	rules v1.SecurityRuleList,
	upstreams gloov1.UpstreamList,
	pods customkube.PodList,
	resourceErrs reporter.ResourceErrors) SecurityConfig {

	rbacConfig := &v1alpha1.RbacConfig{
		Metadata: core.Metadata{
			// required name for global rbac config
			// https://istio.io/docs/concepts/security/#enabling-authorization
			Name:      "default",
			Namespace: writeNamespace,
		},
		Mode: v1alpha1.RbacConfig_OFF,
	}

	// if no security rules are set, disable istio rbac
	if len(rules) == 0 {
		return SecurityConfig{RbacConfig: rbacConfig}
	}

	// enable istio rbac
	rbacConfig.Mode = v1alpha1.RbacConfig_ON
	rbacConfig.EnforcementMode = v1alpha1.EnforcementMode_ENFORCED

	var (
		serviceRoles        v1alpha1.ServiceRoleList
		serviceRoleBindings v1alpha1.ServiceRoleBindingList
	)
	// for each rule:
	for _, r := range rules {
		// create a servicerole for that rule
		sr, err := createServiceRoleFromRule(writeNamespace, r, upstreams)
		if err != nil {
			resourceErrs.AddError(r, err)
			continue
		}
		serviceRoles = append(serviceRoles, sr)

		// create the binding for that role:
		bindingForRole, err := createServiceRoleBinding(
			sr.Metadata.Name,
			sr.Metadata.Namespace,
			r.SourceSelector,
			upstreams,
			pods,
		)
		if err != nil {
			resourceErrs.AddError(r, err)
			continue
		}
		serviceRoleBindings = append(serviceRoleBindings, bindingForRole)
	}

	return SecurityConfig{
		RbacConfig:          rbacConfig,
		ServiceRoles:        serviceRoles,
		ServiceRoleBindings: serviceRoleBindings,
	}
}

func hostsForSelector(selector *v1.PodSelector, upstreams gloov1.UpstreamList) ([]string, error) {
	selectedUpstreams, err := utils.UpstreamsForSelector(selector, upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "selecting upstreams")
	}

	var hostNames []string
	for _, us := range selectedUpstreams {
		hostForUpstream, err := utils.GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}

		hostNames = append(hostNames, hostForUpstream)
	}

	return stringutils.Unique(hostNames), nil
}

func createServiceRoleFromRule(writeNamespace string, rule *v1.SecurityRule, upstreams gloov1.UpstreamList) (*v1alpha1.ServiceRole, error) {
	serviceNames, err := hostsForSelector(rule.DestinationSelector, upstreams)
	if err != nil {
		return nil, err
	}

	allowedPaths := rule.AllowedPaths
	allowedMethods := rule.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"*"}
	}

	return &v1alpha1.ServiceRole{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      rule.Metadata.Namespace + "-" + rule.Metadata.Name,
		},
		Rules: []*v1alpha1.AccessRule{{
			Services: serviceNames,
			Paths:    allowedPaths,
			Methods:  allowedMethods,
		}},
	}, nil
}

func principalName(s core.ResourceRef) string {
	return fmt.Sprintf("cluster.local/ns/%s/sa/%s", s.Namespace, s.Name)
}

func getSubjectsForSelector(selector *v1.PodSelector,
	upstreams gloov1.UpstreamList,
	pods customkube.PodList) ([]*v1alpha1.Subject, error) {
	selectedPods, err := utils.PodsForSelector(selector, upstreams, pods)
	if err != nil {
		return nil, errors.Wrapf(err, "selecting pods")
	}
	var serviceAccounts []core.ResourceRef

	// create a subject for each unique service account
	addSvcAcct := func(newSa core.ResourceRef) {
		for _, sa := range serviceAccounts {
			if sa.Equal(newSa) {
				return
			}
		}
		serviceAccounts = append(serviceAccounts, newSa)
	}
	for _, p := range selectedPods {
		kubePod, err := kubernetes.ToKube(p)
		if err != nil {
			return nil, errors.Wrapf(err, "internal error: converting custom pod object")
		}

		addSvcAcct(core.ResourceRef{
			Name:      kubePod.Spec.ServiceAccountName,
			Namespace: kubePod.Namespace,
		})
	}
	// create a subject for every unique service account
	var subjects []*v1alpha1.Subject
	for _, sa := range serviceAccounts {
		subjects = append(subjects, &v1alpha1.Subject{
			// see example
			// https://istio.io/docs/tasks/security/role-based-access-control/#step-3-allowing-access-to-the-ratings-service
			Properties: map[string]string{
				"source.principal": principalName(sa),
			},
		})
	}

	return subjects, nil
}

func createServiceRoleBinding(
	serviceRoleName string,
	serviceRoleNamespace string,
	sourceSelector *v1.PodSelector,
	upstreams gloov1.UpstreamList,
	pods customkube.PodList) (*v1alpha1.ServiceRoleBinding, error) {

	subjects, err := getSubjectsForSelector(sourceSelector, upstreams, pods)
	if err != nil {
		return nil, errors.Wrapf(err, "finding subjects (service accounts) for source selector")
	}

	return &v1alpha1.ServiceRoleBinding{
		Metadata: core.Metadata{
			Name:      serviceRoleName,
			Namespace: serviceRoleNamespace,
		},
		Subjects: subjects,
		RoleRef: &v1alpha1.RoleRef{
			Kind: "ServiceRole",
			Name: serviceRoleName,
		},
	}, nil
}
