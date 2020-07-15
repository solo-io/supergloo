package istio

import (
	"context"
	"fmt"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		in input.Snapshot,
		reporter reporter.Reporter,
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
	reporter reporter.Reporter,
) (istio.Snapshot, error) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	meshServiceTranslator := t.dependencies.makeMeshServiceTranslator(in.KubernetesClusters())

	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	virtualServices := v1alpha3sets.NewVirtualServiceSet()
	envoyFilters := v1alpha3sets.NewEnvoyFilterSet()
	gateways := v1alpha3sets.NewGatewaySet()
	serviceEntries := v1alpha3sets.NewServiceEntrySet()

	for _, meshService := range in.MeshServices().List() {
		meshService := meshService // pike

		virtualService, destinationRule := meshServiceTranslator.Translate(in, meshService, reporter)

		if destinationRule != nil {
			destinationRules.Insert(destinationRule)
			contextutils.LoggerFrom(ctx).Debugf("translated destination rule %v", sets.Key(meshService))
		}
		if virtualService != nil {
			contextutils.LoggerFrom(ctx).Debugf("translated virtual service %v", sets.Key(meshService))
			virtualServices.Insert(virtualService)
		}
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
