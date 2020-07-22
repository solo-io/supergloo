package authorizationpolicy

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discovery_smh_solo_io_v1alpha2_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/ezkube"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	typesv1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

const (
	translatorName = "authorization-policy-translator"
)

var (
	trustDomainNotFound = func(clusterName string) error {
		return eris.Errorf("Trust domain not found for cluster %s", clusterName)
	}
)

// the AuthorizationPolicy translator translates a MeshService into a AuthorizationPolicy.
type Translator interface {
	// Translate translates an appropriate AuthorizationPolicy for the given MeshService.
	// returns nil if no AuthorizationPolicy is required for the MeshService (i.e. if no AuthorizationPolicy features are required, such access control).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot MeshServiceSet contains the given MeshService.
	Translate(
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		reporter reporting.Reporter,
	) *securityv1beta1.AuthorizationPolicy
}

type translator struct {
	decoratorFactory decorators.Factory
}

func NewTranslator(decoratorFactory decorators.Factory) *translator {
	return &translator{decoratorFactory: decoratorFactory}
}

func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	reporter reporting.Reporter,
) *securityv1beta1.AuthorizationPolicy {
	kubeService := meshService.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(harveyxia): non kube services currently unsupported
		return nil
	}

	authPolicy := t.initializeAuthorizationPolicy(meshService)
	// register the owners of the AuthorizationPolicy fields
	authPolicyFields := fieldutils.NewOwnershipRegistry()
	apDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		Snapshot: in,
	})

	// Apply decorators which map each AccessPolicy to a AuthorizationPolicy Rule.
	for _, policy := range meshService.Status.AppliedAccessPolicies {
		baseRule, err := t.initializeBaseRule(policy.Spec, in.Meshes())
		if err != nil {
			// If error encountered while translating source selector, do not translate.
			reporter.ReportAccessPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", translatorName))
			continue
		}

		registerField := registerFieldFunc(authPolicyFields, authPolicy, []ezkube.ResourceId{policy.Ref})
		for _, decorator := range apDecorators {
			if authPolicyDecorator, ok := decorator.(accesspolicy.AuthorizationPolicyDecorator); ok {
				if err := authPolicyDecorator.ApplyToAuthorizationPolicy(
					policy,
					meshService,
					baseRule.To[0].Operation,
					registerField,
				); err != nil {
					reporter.ReportAccessPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
				}
			}
		}

		authPolicy.Spec.Rules = append(authPolicy.Spec.Rules, baseRule)
	}

	if len(authPolicy.Spec.Rules) == 0 {
		// no need to create this AuthorizationPolicy as it has no effect
		return nil
	}
	return authPolicy
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	authPolicyFields fieldutils.FieldOwnershipRegistry,
	authPolicy *securityv1beta1.AuthorizationPolicy,
	policyRefs []ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.Equals(fieldVal, val) {
			return nil
		}
		if err := authPolicyFields.RegisterFieldOwnership(
			authPolicy,
			fieldPtr,
			policyRefs,
			&v1alpha2.AccessPolicy{},
			0, //TODO(harveyxia): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeAuthorizationPolicy(meshService *discoveryv1alpha2.MeshService) *securityv1beta1.AuthorizationPolicy {
	meta := metautils.TranslatedObjectMeta(
		meshService.Spec.GetKubeService().Ref,
		meshService.Annotations,
	)
	authPolicy := &securityv1beta1.AuthorizationPolicy{
		ObjectMeta: meta,
		Spec: securityv1beta1spec.AuthorizationPolicy{
			Selector: &typesv1beta1.WorkloadSelector{
				MatchLabels: meshService.Spec.GetKubeService().WorkloadSelectorLabels,
			},
			Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
		},
	}
	return authPolicy
}

// Initialize AccessPolicy Rule consisting of a Rule_From for each SourceSelector and a single Rule_To containing the rules specified in the AccessPolicy.
func (t *translator) initializeBaseRule(
	accessPolicy *v1alpha2.AccessPolicySpec,
	meshes discovery_smh_solo_io_v1alpha2_sets.MeshSet,
) (*securityv1beta1spec.Rule, error) {
	var fromRules []*securityv1beta1spec.Rule_From
	for _, sourceSelector := range accessPolicy.SourceSelector {
		fromRule, err := t.buildSource(sourceSelector, meshes)
		if err != nil {
			return nil, err
		}
		fromRules = append(fromRules, fromRule)
	}
	return &securityv1beta1spec.Rule{
		From: fromRules,
		To: []*securityv1beta1spec.Rule_To{
			{
				Operation: &securityv1beta1spec.Operation{}, // To be populated by AuthorizationPolicyDecorators
			},
		},
	}, nil
}

// Generate all fully qualified principal names for specified service accounts.
// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#Source
func (t *translator) buildSource(
	sources *v1alpha2.IdentitySelector,
	meshes discovery_smh_solo_io_v1alpha2_sets.MeshSet,
) (*securityv1beta1spec.Rule_From, error) {
	var principals []string
	var namespaces []string
	if sources.GetKubeIdentityMatcher() == nil && sources.GetKubeServiceAccountRefs() == nil {
		// allow any source identity
		return &securityv1beta1spec.Rule_From{
			Source: &securityv1beta1spec.Source{},
		}, nil
	}
	// Process Matcher
	if len(sources.GetKubeIdentityMatcher().GetClusters()) > 0 {
		// select by clusters and specifiedNamespaces
		trustDomains, err := t.getTrustDomainsForClusters(sources.GetKubeIdentityMatcher().GetClusters(), meshes)
		if err != nil {
			return nil, err
		}
		specifiedNamespaces := sources.GetKubeIdentityMatcher().GetNamespaces()
		// Permit any namespace if unspecified.
		if len(specifiedNamespaces) == 0 {
			specifiedNamespaces = []string{""}
		}
		for _, trustDomain := range trustDomains {
			for _, namespace := range specifiedNamespaces {
				// Use empty string for service account to permit any.
				uri, err := buildSpiffeURI(trustDomain, namespace, "")
				if err != nil {
					return nil, err
				}
				principals = append(principals, uri)
			}
		}
	} else if len(sources.GetKubeIdentityMatcher().GetNamespaces()) > 0 {
		// select by namespaces, permit any cluster
		namespaces = sources.GetKubeIdentityMatcher().GetNamespaces()
	}

	// select by direct reference to ServiceAccounts
	for _, serviceAccountRef := range sources.GetKubeServiceAccountRefs().GetServiceAccounts() {
		trustDomains, err := t.getTrustDomainsForClusters([]string{serviceAccountRef.ClusterName}, meshes)
		if err != nil {
			return nil, err
		}
		// If no error thrown, trustDomains is guaranteed to be of length 1.
		uri, err := buildSpiffeURI(trustDomains[0], serviceAccountRef.Namespace, serviceAccountRef.Name)
		if err != nil {
			return nil, err
		}
		principals = append(principals, uri)
	}

	return &securityv1beta1spec.Rule_From{
		Source: &securityv1beta1spec.Source{
			Principals: principals,
			Namespaces: namespaces,
		},
	}, nil
}

/*
	Fetch trust domains for the Istio mesh of the given cluster.
	Multiple mesh installations of the same type on the same cluster are unsupported, simply use the first Mesh encountered.
*/
func (t *translator) getTrustDomainsForClusters(
	clusterNames []string,
	meshes discovery_smh_solo_io_v1alpha2_sets.MeshSet,
) ([]string, error) {
	var errs *multierror.Error
	var trustDomains []string
	for _, clusterName := range clusterNames {
		trustDomain, err := t.getTrustDomainForCluster(clusterName, meshes)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		trustDomains = append(trustDomains, trustDomain)
	}
	return trustDomains, errs.ErrorOrNil()
}

// Fetch trust domains by cluster so we can attribute missing trust domains to the problematic clusterName and report back to user.
func (t *translator) getTrustDomainForCluster(
	clusterName string,
	meshes discovery_smh_solo_io_v1alpha2_sets.MeshSet,
) (string, error) {
	var trustDomain string
	for _, mesh := range meshes.List(func(mesh *discoveryv1alpha2.Mesh) bool {
		return mesh.Spec.GetIstio() == nil || mesh.Spec.GetIstio().GetInstallation().GetCluster() != clusterName
	}) {
		trustDomain = mesh.Spec.GetIstio().GetCitadelInfo().GetTrustDomain()
	}
	if trustDomain == "" {
		return "", trustDomainNotFound(clusterName)
	}
	return trustDomain, nil
}

/*
	The principal string format is described here: https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#2-spiffe-identity
	Testing shows that the "spiffe://" prefix cannot be included.
	Istio only respects prefix or suffix wildcards, https://github.com/istio/istio/blob/9727308b3dadbfc8151cf70a045d1c7c52ab222b/pilot/pkg/security/authz/model/matcher/string.go#L45
*/
func buildSpiffeURI(trustDomain, namespace, serviceAccount string) (string, error) {
	if trustDomain == "" {
		return "", eris.New("trustDomain cannot be empty")
	}
	if namespace == "" {
		return fmt.Sprintf("%s/ns/*", trustDomain), nil
	} else if serviceAccount == "" {
		return fmt.Sprintf("%s/ns/%s/sa/*", trustDomain, namespace), nil
	} else {
		return fmt.Sprintf("%s/ns/%s/sa/%s", trustDomain, namespace, serviceAccount), nil
	}
}
