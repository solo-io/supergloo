package istio_translator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	istio_security "github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	access_control_policy "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-policy-translator"
	security_v1beta1 "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	client_security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TranslatorId = "istio-translator"
)

var (
	ACPProcessingError = func(err error, acp *networking_v1alpha1.AccessControlPolicy) error {
		return eris.Wrapf(err, "Error processing AccessControlPolicy: %+v", acp)
	}
	AuthPolicyUpsertError = func(err error, authPolicy *client_security_v1beta1.AuthorizationPolicy) error {
		return eris.Wrapf(err, "Error while upserting AuthorizationPolicy: %+v", authPolicy)
	}
	EmptyTrustDomainForMeshError = func(mesh *discovery_v1alpha1.Mesh) error {
		return eris.Errorf("Empty trust domain for Istio Mesh: %+v", mesh)
	}
	EmptyTrustDomainForCluster = func(clusterName string) error {
		return eris.Errorf("No trust domain found for cluster: %s", clusterName)
	}
)

type IstioTranslator access_control_policy.AcpMeshTranslator

type istioTranslator struct {
	authPolicyClientFactory security.AuthorizationPolicyClientFactory
	meshClient              zephyr_discovery.MeshClient
	dynamicClientGetter     mc_manager.DynamicClientGetter
}

func NewIstioTranslator(
	meshClient zephyr_discovery.MeshClient,
	dynamicClientGetter mc_manager.DynamicClientGetter,
	authPolicyClientFactory security.AuthorizationPolicyClientFactory,
) IstioTranslator {
	return &istioTranslator{
		authPolicyClientFactory: authPolicyClientFactory,
		meshClient:              meshClient,
		dynamicClientGetter:     dynamicClientGetter,
	}
}

/*
Translate AccessControlPolicy into an Istio AuthorizationPolicy and upsert it into the cluster of the target k8s Service.
Compute all AuthorizationPolicies for the given target Services before performing any upserts in order to preserve atomicity
in the case of processing errors.
*/
func (i *istioTranslator) Translate(
	ctx context.Context,
	targetServices []access_control_policy.TargetService,
	acp *networking_v1alpha1.AccessControlPolicy,
) *networking_types.AccessControlPolicyStatus_TranslatorError {
	authPoliciesWithClients := make([]authPolicyClientPair, 0, len(targetServices))
	fromSources, err := i.buildSources(ctx, acp.Spec.GetSourceSelector())
	if err != nil {
		return &networking_types.AccessControlPolicyStatus_TranslatorError{
			TranslatorId: TranslatorId,
			ErrorMessage: ACPProcessingError(err, acp).Error(),
		}
	}
	// First compute all AuthorizationPolicies for each target k8s Service, and then perform all upserts.
	// This ensures that the AccessControlPolicy is only applied (i.e. Istio config created) if the translation succeeds for all
	// target services.
	for _, targetService := range targetServices {
		// only translate Istio-backed services
		if targetService.Mesh.Spec.GetIstio() == nil {
			continue
		}
		client, err := i.dynamicClientGetter.GetClientForCluster(targetService.Mesh.Spec.GetCluster().GetName())
		if err != nil {
			return &networking_types.AccessControlPolicyStatus_TranslatorError{
				TranslatorId: TranslatorId,
				ErrorMessage: err.Error(),
			}
		}
		authPolicyWithClient := authPolicyClientPair{
			authPolicy: i.translateForDestination(fromSources, acp, targetService.MeshService),
			client:     i.authPolicyClientFactory(client),
		}
		authPoliciesWithClients = append(authPoliciesWithClients, authPolicyWithClient)
	}
	// upsert all computed AuthorizationPolicies
	for _, authPolicyWithClient := range authPoliciesWithClients {
		err := authPolicyWithClient.client.UpsertSpec(ctx, authPolicyWithClient.authPolicy)
		if err != nil {
			return &networking_types.AccessControlPolicyStatus_TranslatorError{
				TranslatorId: TranslatorId,
				ErrorMessage: AuthPolicyUpsertError(err, authPolicyWithClient.authPolicy).Error(),
			}
		}
	}
	return nil
}

func (i *istioTranslator) translateForDestination(
	fromSources *security_v1beta1.Rule_From,
	acp *networking_v1alpha1.AccessControlPolicy,
	meshService *discovery_v1alpha1.MeshService,
) *client_security_v1beta1.AuthorizationPolicy {
	allowedMethods := methodsToString(acp.Spec.GetAllowedMethods())
	// Istio considers AuthorizationPolicies without at least one defined To.Operation invalid,
	// The workaround is to populate a dummy "*" for METHOD if not user specified. This guarantees existence of at least
	// one To.Operation.
	if len(allowedMethods) < 1 {
		allowedMethods = []string{"*"}
	}
	authPolicy := &client_security_v1beta1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      acp.GetName(),
			Namespace: meshService.Spec.GetKubeService().GetRef().GetNamespace(),
		},
		Spec: security_v1beta1.AuthorizationPolicy{
			Selector: &v1beta1.WorkloadSelector{
				MatchLabels: meshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
			},
			Rules: []*security_v1beta1.Rule{
				{
					From: []*security_v1beta1.Rule_From{
						fromSources,
					},
					To: []*security_v1beta1.Rule_To{
						{
							Operation: &security_v1beta1.Operation{
								Ports:   intToString(acp.Spec.GetAllowedPorts()),
								Methods: allowedMethods,
								Paths:   acp.Spec.GetAllowedPaths(),
							},
						},
					},
				},
			},
			Action: security_v1beta1.AuthorizationPolicy_ALLOW,
		},
	}
	return authPolicy
}

// Generate all fully qualified principal names for specified service accounts.
// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#Source
func (i *istioTranslator) buildSources(
	ctx context.Context,
	source *core_types.IdentitySelector,
) (*security_v1beta1.Rule_From, error) {
	var principals []string
	var namespaces []string
	if source.GetIdentitySelectorType() == nil {
		// allow any source identity
		return &security_v1beta1.Rule_From{
			Source: &security_v1beta1.Source{
				Principals: []string{"*"},
			},
		}, nil
	}
	switch selectorType := source.GetIdentitySelectorType().(type) {
	case *core_types.IdentitySelector_Matcher_:
		if len(source.GetMatcher().GetClusters()) > 0 {
			// select by clusters and specifiedNamespaces
			trustDomains, err := i.getTrustDomainForClusters(ctx, source.GetMatcher().GetClusters())
			if err != nil {
				return nil, err
			}
			specifiedNamespaces := source.GetMatcher().GetNamespaces()
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
		} else {
			// select by namespaces, permit any cluster
			namespaces = source.GetMatcher().GetNamespaces()
		}
	case *core_types.IdentitySelector_ServiceAccountRefs_:
		// select by direct reference to ServiceAccounts
		for _, serviceAccountRef := range source.GetServiceAccountRefs().GetServiceAccounts() {
			trustDomains, err := i.getTrustDomainForClusters(ctx, []string{serviceAccountRef.GetCluster()})
			if err != nil {
				return nil, err
			}
			// If no error thrown, trustDomains is guaranteed to be of length 1.
			uri, err := buildSpiffeURI(trustDomains[0], serviceAccountRef.GetNamespace(), serviceAccountRef.GetName())
			if err != nil {
				return nil, err
			}
			principals = append(principals, uri)
		}
	default:
		return nil, eris.Errorf("IdentitySelector has unexpected type %T", selectorType)
	}
	return &security_v1beta1.Rule_From{
		Source: &security_v1beta1.Source{
			Principals: principals,
			Namespaces: namespaces,
		},
	}, nil
}

/*
Fetch trust domains for the Istio mesh of the given cluster.
Multiple mesh installations of the same type on the same cluster are unsupported, simply use the first Mesh encountered.
*/
func (i *istioTranslator) getTrustDomainForClusters(ctx context.Context, clusterNames []string) ([]string, error) {
	meshList, err := i.meshClient.List(ctx)
	if err != nil {
		return nil, err
	}
	var trustDomains []string
	for _, mesh := range meshList.Items {
		mesh := mesh
		if mesh.Spec.GetIstio() == nil || !stringutils.ContainsString(mesh.Spec.GetCluster().GetName(), clusterNames) {
			continue
		}
		trustDomain := mesh.Spec.GetIstio().GetCitadelInfo().GetTrustDomain()
		if trustDomain == "" {
			return nil, EmptyTrustDomainForMeshError(&mesh)
		}
		trustDomains = append(trustDomains, trustDomain)
	}
	return trustDomains, nil
}

func methodsToString(methodEnums []core_types.HttpMethodValue) []string {
	var methods []string
	for _, methodEnum := range methodEnums {
		methods = append(methods, methodEnum.String())
	}
	return methods
}

// Pair of AuthorizationPolicy with client scoped to the target k8s Service's cluster
type authPolicyClientPair struct {
	authPolicy *client_security_v1beta1.AuthorizationPolicy
	client     istio_security.AuthorizationPolicyClient
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

func intToString(ints []uint32) []string {
	var strings []string
	for _, i := range ints {
		strings = append(strings, strconv.Itoa(int(i)))
	}
	return strings
}
