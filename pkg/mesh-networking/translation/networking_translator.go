package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"

	certificatesv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2/sets"

	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	v1beta1sets "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
)

// the networking translator translates an input networking snapshot to an output snapshot of mesh config resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		in input.Snapshot,
		reporter reporting.Reporter,
	) (output.Snapshot, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	istioTranslator istio.Translator
}

func NewTranslator(
	istioTranslator istio.Translator,
) Translator {
	return &translator{
		istioTranslator: istioTranslator,
	}
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	reporter reporting.Reporter,
) (output.Snapshot, error) {
	t.totalTranslates++
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("translation-%v", t.totalTranslates))

	destinationRules := v1alpha3sets.NewDestinationRuleSet()
	virtualServices := v1alpha3sets.NewVirtualServiceSet()
	authorizationPolicies := v1beta1sets.NewAuthorizationPolicySet()
	envoyFilters := v1alpha3sets.NewEnvoyFilterSet()
	gateways := v1alpha3sets.NewGatewaySet()
	serviceEntries := v1alpha3sets.NewServiceEntrySet()
	issuedCertificates := certificatesv1alpha2sets.NewIssuedCertificateSet()

	istioOutputs := t.istioTranslator.Translate(ctx, in, reporter)

	destinationRules = destinationRules.Union(istioOutputs.DestinationRules)
	virtualServices = virtualServices.Union(istioOutputs.VirtualServices)
	authorizationPolicies = authorizationPolicies.Union(istioOutputs.AuthorizationPolicies)
	envoyFilters = envoyFilters.Union(istioOutputs.EnvoyFilters)
	gateways = gateways.Union(istioOutputs.Gateways)
	serviceEntries = serviceEntries.Union(istioOutputs.ServiceEntries)
	issuedCertificates = issuedCertificates.Union(istioOutputs.IssuedCertificates)

	return output.NewSinglePartitionedSnapshot(
		fmt.Sprintf("networking-%v", t.totalTranslates),
		metautils.TranslatedObjectLabels(),
		issuedCertificates,
		destinationRules,
		envoyFilters,
		gateways,
		serviceEntries,
		virtualServices,
		authorizationPolicies,
	)
}
