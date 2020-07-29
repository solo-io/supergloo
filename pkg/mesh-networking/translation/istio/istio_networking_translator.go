package istio

import (
	"context"
	"fmt"

	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	v1beta1sets "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/internal"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		in input.Snapshot,
		reporter reporting.Reporter,
	) (istio.Snapshot, error)
}

type istioTranslator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    internal.DependencyFactory
}

func NewIstioTranslator() Translator {
	return &istioTranslator{
		dependencies: internal.NewDependencyFactory(),
	}
}

func (t *istioTranslator) Translate(
	ctx context.Context,
	in input.Snapshot,
	reporter reporting.Reporter,
) (istio.Snapshot, error) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	meshServiceTranslator := t.dependencies.MakeMeshServiceTranslator(in.KubernetesClusters())

	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	virtualServices := v1alpha3sets.NewVirtualServiceSet()
	authorizationPolicies := v1beta1sets.NewAuthorizationPolicySet()

	for _, meshService := range in.MeshServices().List() {
		meshService := meshService // pike

		serviceOutputs := meshServiceTranslator.Translate(in, meshService, reporter)

		destinationRule := serviceOutputs.DestinationRule
		if destinationRule != nil {
			destinationRules.Insert(destinationRule)
			contextutils.LoggerFrom(ctx).Debugf("translated destination rule %v", sets.Key(destinationRule))
		}

		virtualService := serviceOutputs.VirtualService
		if virtualService != nil {
			contextutils.LoggerFrom(ctx).Debugf("translated virtual service %v", sets.Key(virtualService))
			virtualServices.Insert(virtualService)
		}

		authorizationPolicy := serviceOutputs.AuthorizationPolicy
		if authorizationPolicy != nil {
			contextutils.LoggerFrom(ctx).Debugf("translated authorization policy %v", sets.Key(authorizationPolicy))
			authorizationPolicies.Insert(authorizationPolicy)
		}
	}

	envoyFilters := v1alpha3sets.NewEnvoyFilterSet()
	gateways := v1alpha3sets.NewGatewaySet()
	serviceEntries := v1alpha3sets.NewServiceEntrySet()

	meshTranslator := t.dependencies.MakeMeshTranslator(ctx, in.KubernetesClusters())
	for _, mesh := range in.Meshes().List() {
		meshOutputs := meshTranslator.Translate(in, mesh, reporter)

		gateways = gateways.Union(meshOutputs.Gateways)
		serviceEntries = serviceEntries.Union(meshOutputs.ServiceEntries)
		envoyFilters = envoyFilters.Union(meshOutputs.EnvoyFilters)
		destinationRules = destinationRules.Union(meshOutputs.DestinationRules)
		authorizationPolicies = authorizationPolicies.Union(meshOutputs.AuthorizationPolicies)
	}

	t.totalTranslates++

	return istio.NewSinglePartitionedSnapshot(
		fmt.Sprintf("istio-networking-%v", t.totalTranslates),
		metautils.TranslatedObjectLabels(),
		destinationRules,
		envoyFilters,
		gateways,
		serviceEntries,
		virtualServices,
		authorizationPolicies,
	)
}
