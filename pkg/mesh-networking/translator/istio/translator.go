package istio

import (
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/istio/destinationrule"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/utils/metautils"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		in input.Snapshot,
		reporter reporter.Reporter,
	) (output.Snapshot, error)
}

type translator struct {
	destinationRules destinationrule.Translator
}

func (t *translator) Translate(
	in input.Snapshot,
	reporter reporter.Reporter,
) (output.Snapshot, error) {
	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	for _, meshService := range in.MeshServices().List() {
		destinationRule := t.destinationRules.Translate(in, meshService, reporter)
		if destinationRule == nil {
			continue
		}
		destinationRules.Insert(destinationRule)
	}

	virtualServices := v1alpha3sets.NewVirtualServiceSet()
	envoyFilters := v1alpha3sets.NewEnvoyFilterSet()

	return output.NewLabelPartitionedSnapshot(
		"istio-networking",
		metautils.OwnershipLabelKey,
		destinationRules,
		envoyFilters,
		virtualServices,
	)
}
