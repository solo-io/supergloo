package istio

import (
	"context"
	"fmt"

	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/reporting"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
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
	dependencies    dependencyFactory
}

func NewIstioTranslator() Translator {
	return &istioTranslator{
		dependencies: dependencyFactoryImpl{},
	}
}

func (t *istioTranslator) Translate(
	ctx context.Context,
	in input.Snapshot,
	reporter reporting.Reporter,
) (istio.Snapshot, error) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	meshServiceTranslator := t.dependencies.makeMeshServiceTranslator(in.KubernetesClusters())

	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	virtualServices := v1alpha3sets.NewVirtualServiceSet()

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
	}

	envoyFilters := v1alpha3sets.NewEnvoyFilterSet()
	gateways := v1alpha3sets.NewGatewaySet()
	serviceEntries := v1alpha3sets.NewServiceEntrySet()

	meshTranslator := t.dependencies.makeMeshTranslator(ctx, in.KubernetesClusters())
	for _, mesh := range in.Meshes().List() {
		meshOutputs := meshTranslator.Translate(in, mesh, reporter)

		gateways = gateways.Union(meshOutputs.Gateways)
		serviceEntries = serviceEntries.Union(meshOutputs.ServiceEntries)
		envoyFilters = envoyFilters.Union(meshOutputs.EnvoyFilters)
		destinationRules = destinationRules.Union(meshOutputs.DestinationRules)
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
	)
}
