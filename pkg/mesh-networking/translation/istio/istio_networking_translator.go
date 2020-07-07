package istio

import (
	"fmt"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	"github.com/solo-io/smh/pkg/mesh-networking/translation"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/destinationrule"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/virtualservice"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		in translation.Snapshot,
		reporter reporter.Reporter,
	) (istio.Snapshot, error)
}

type istioTranslator struct {
	totalTranslates  int // TODO(ilackarms): metric
	destinationRules destinationrule.Translator
	virtualServices  virtualservice.Translator
}

func (t *istioTranslator) Translate(
	in translation.Snapshot,
	reporter reporter.Reporter,
) (istio.Snapshot, error) {
	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	virtualServices := v1alpha3sets.NewVirtualServiceSet()
	for _, meshService := range in.MeshServices().List() {
		destinationRule := t.destinationRules.Translate(in, meshService, reporter)
		if destinationRule != nil {
			destinationRules.Insert(destinationRule)
		}
		virtualService := t.virtualServices.Translate(in, meshService, reporter)
		if virtualService != nil {
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
