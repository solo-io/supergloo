package istio_translator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	istio_security "github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	access_control_policy "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator"
	istio_security_types "istio.io/api/security/v1beta1"
	istio_types "istio.io/api/type/v1beta1"
	istio_security_client_types "istio.io/client-go/pkg/apis/security/v1beta1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TranslatorId = "istio_translator"
)

var (
	ACPProcessingError = func(err error, acp *zephyr_networking.AccessControlPolicy) error {
		return eris.Wrapf(err, "Error processing AccessControlPolicy: %s.%s", acp.GetName(), acp.GetNamespace())
	}
	AuthPolicyUpsertError = func(err error, authPolicy *istio_security_client_types.AuthorizationPolicy) error {
		return eris.Wrapf(err, "Error while upserting AuthorizationPolicy: %s.%s",
			authPolicy.Name, authPolicy.Namespace)
	}
	EmptyTrustDomainForMeshError = func(mesh *zephyr_discovery.Mesh, info *zephyr_discovery_types.MeshSpec_IstioMesh_CitadelInfo) error {
		return eris.Errorf("Empty trust domain for Istio Mesh: %s.%s, (%+v)", mesh.Name, mesh.Namespace, info)
	}
	ServiceAccountRefNonexistent = func(saRef *zephyr_core_types.ResourceRef) error {
		return eris.Errorf("Service account ref did not match a real service account: %s.%s.%s", saRef.Name, saRef.Namespace, saRef.Cluster)
	}
)

type IstioTranslator access_control_policy.AcpMeshTranslator

type istioTranslator struct {
	authPolicyClientFactory istio_security.AuthorizationPolicyClientFactory
	meshClient              zephyr_discovery.MeshClient
	dynamicClientGetter     mc_manager.DynamicClientGetter
}

func NewIstioTranslator(
	meshClient zephyr_discovery.MeshClient,
	dynamicClientGetter mc_manager.DynamicClientGetter,
	authPolicyClientFactory istio_security.AuthorizationPolicyClientFactory,
) IstioTranslator {
	return &istioTranslator{
		authPolicyClientFactory: authPolicyClientFactory,
		meshClient:              meshClient,
		dynamicClientGetter:     dynamicClientGetter,
	}
}

func (i *istioTranslator) Name() string {
	return TranslatorId
}

/*
Translate AccessControlPolicy into an Istio AuthorizationPolicy and upsert it into the cluster of the target k8s Service.
Compute all AuthorizationPolicies for the given target Services before performing any upserts in order to preserve atomicity
in the case of processing errors.
*/
func (i *istioTranslator) Translate(
	ctx context.Context,
	targetServices []access_control_policy.TargetService,
	acp *zephyr_networking.AccessControlPolicy,
) *zephyr_networking_types.AccessControlPolicyStatus_TranslatorError {
	authPoliciesWithClients := make([]authPolicyClientPair, 0, len(targetServices))
	fromSources, err := i.buildSources(ctx, acp.Spec.GetSourceSelector())
	if err != nil {
		return &zephyr_networking_types.AccessControlPolicyStatus_TranslatorError{
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
		client, err := i.dynamicClientGetter.GetClientForCluster(ctx, targetService.Mesh.Spec.GetCluster().GetName())
		if err != nil {
			return &zephyr_networking_types.AccessControlPolicyStatus_TranslatorError{
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
		err := authPolicyWithClient.client.UpsertAuthorizationPolicySpec(ctx, authPolicyWithClient.authPolicy)
		if err != nil {
			return &zephyr_networking_types.AccessControlPolicyStatus_TranslatorError{
				TranslatorId: TranslatorId,
				ErrorMessage: AuthPolicyUpsertError(err, authPolicyWithClient.authPolicy).Error(),
			}
		}
	}
	return nil
}

func (i *istioTranslator) translateForDestination(
	fromSources *istio_security_types.Rule_From,
	acp *zephyr_networking.AccessControlPolicy,
	meshService *zephyr_discovery.MeshService,
) *istio_security_client_types.AuthorizationPolicy {
	allowedMethods := methodsToString(acp.Spec.GetAllowedMethods())
	// Istio considers AuthorizationPolicies without at least one defined To.Operation invalid,
	// The workaround is to populate a dummy "*" for METHOD if not user specified. This guarantees existence of at least
	// one To.Operation.
	if len(allowedMethods) < 1 {
		allowedMethods = []string{"*"}
	}
	authPolicy := &istio_security_client_types.AuthorizationPolicy{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      buildAuthPolicyName(acp, meshService),
			Namespace: meshService.Spec.GetKubeService().GetRef().GetNamespace(),
		},
		Spec: istio_security_types.AuthorizationPolicy{
			Selector: &istio_types.WorkloadSelector{
				MatchLabels: meshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
			},
			Rules: []*istio_security_types.Rule{
				{
					From: []*istio_security_types.Rule_From{
						fromSources,
					},
					To: []*istio_security_types.Rule_To{
						{
							Operation: &istio_security_types.Operation{
								Ports:   intToString(acp.Spec.GetAllowedPorts()),
								Methods: allowedMethods,
								Paths:   acp.Spec.GetAllowedPaths(),
							},
						},
					},
				},
			},
			Action: istio_security_types.AuthorizationPolicy_ALLOW,
		},
	}
	return authPolicy
}

// Generate all fully qualified principal names for specified service accounts.
// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#Source
func (i *istioTranslator) buildSources(
	ctx context.Context,
	source *zephyr_core_types.IdentitySelector,
) (*istio_security_types.Rule_From, error) {
	var principals []string
	var namespaces []string
	if source.GetIdentitySelectorType() == nil {
		// allow any source identity
		return &istio_security_types.Rule_From{
			Source: &istio_security_types.Source{
				Principals: []string{"*"},
			},
		}, nil
	}
	switch selectorType := source.GetIdentitySelectorType().(type) {
	case *zephyr_core_types.IdentitySelector_Matcher_:
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
	case *zephyr_core_types.IdentitySelector_ServiceAccountRefs_:
		// select by direct reference to ServiceAccounts
		for _, serviceAccountRef := range source.GetServiceAccountRefs().GetServiceAccounts() {
			trustDomains, err := i.getTrustDomainForClusters(ctx, []string{serviceAccountRef.GetCluster()})
			if err != nil {
				return nil, err
			}
			if len(trustDomains) == 0 {
				return nil, ServiceAccountRefNonexistent(serviceAccountRef)
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
	return &istio_security_types.Rule_From{
		Source: &istio_security_types.Source{
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
	meshList, err := i.meshClient.ListMesh(ctx)
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
			return nil, EmptyTrustDomainForMeshError(&mesh, mesh.Spec.GetIstio().GetCitadelInfo())
		}
		trustDomains = append(trustDomains, trustDomain)
	}
	return trustDomains, nil
}

func methodsToString(methodEnums []zephyr_core_types.HttpMethodValue) []string {
	var methods []string
	for _, methodEnum := range methodEnums {
		methods = append(methods, methodEnum.String())
	}
	return methods
}

// Pair of AuthorizationPolicy with client scoped to the target k8s Service's cluster
type authPolicyClientPair struct {
	authPolicy *istio_security_client_types.AuthorizationPolicy
	client     istio_security.AuthorizationPolicyClient
}

func buildAuthPolicyName(acp *zephyr_networking.AccessControlPolicy, svc *zephyr_discovery.MeshService) string {
	return fmt.Sprintf("%s-%s", acp.GetName(), svc.GetName())
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
