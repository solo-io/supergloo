package istio

import (
	"context"
	"fmt"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		in input.Snapshot,
		reporter reporter.Reporter,
	) (istio.Snapshot, error)
}

type istioTranslator struct {
	ctx             context.Context
	totalTranslates int // TODO(ilackarms): metric
	dependencies    dependencyFactory
}

func NewIstioTranslator(ctx context.Context) Translator {
	return &istioTranslator{
		ctx:          ctx,
		dependencies: dependencyFactoryImpl{},
	}
}

func (t *istioTranslator) Translate(
	in input.Snapshot,
	reporter reporter.Reporter,
) (istio.Snapshot, error) {

	destinationRuleTranslator := t.dependencies.makeDestinationRuleTranslator(in.KubernetesClusters())

	virtualServiceTranslator := t.dependencies.makeVirtualServiceTranslator(in.KubernetesClusters())

	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	virtualServices := v1alpha3sets.NewVirtualServiceSet()
	for _, meshService := range in.MeshServices().List() {
		destinationRule := destinationRuleTranslator.Translate(in, meshService, reporter)
		if destinationRule != nil {
			destinationRules.Insert(destinationRule)
			contextutils.LoggerFrom(t.ctx).Debugw("translated destination rule %v", sets.Key(meshService))
		}
		virtualService := virtualServiceTranslator.Translate(in, meshService, reporter)
		if virtualService != nil {
			contextutils.LoggerFrom(t.ctx).Debugw("translated virtual service %v", sets.Key(meshService))
			virtualServices.Insert(virtualService)
		}
	}

	envoyFilters := v1alpha3sets.NewEnvoyFilterSet()

	t.totalTranslates++

	return istio.NewLabelPartitionedSnapshot(
		fmt.Sprintf("istio-networking-%v", t.totalTranslates),
		metautils.OwnershipLabelKey,
		destinationRules,
		envoyFilters,
		virtualServices,
	)
}
