package istio

import (
	"context"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
)

func (t *translator) makeDestinatioRuleForHost(
	ctx context.Context,
	params plugins.Params,
	host string,
	labelSets []map[string]string,
	enableMtls bool,
	resourceErrs reporter.ResourceErrors,
) *v1alpha3.DestinationRule {
	dr := initDestinationRule(t.writeNamespace, host, labelSets, enableMtls)

	return dr
}

func initDestinationRule(writeNamespace, host string, labelSets []map[string]string, enableMtls bool) *v1alpha3.DestinationRule {

	var subsets []*v1alpha3.Subset
	for _, labels := range labelSets {
		if len(labels) == 0 {
			continue
		}
		subsets = append(subsets, &v1alpha3.Subset{
			Name:   subsetName(labels),
			Labels: labels,
		})
	}
	var trafficPolicy *v1alpha3.TrafficPolicy
	if enableMtls {
		trafficPolicy = &v1alpha3.TrafficPolicy{
			Tls: &v1alpha3.TLSSettings{
				Mode: v1alpha3.TLSSettings_ISTIO_MUTUAL, // plain old mutual ain't supported yet
			},
		}
	}
	return &v1alpha3.DestinationRule{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      host,
		},
		Host:          host,
		Subsets:       subsets,
		TrafficPolicy: trafficPolicy,
	}
}

func subsetName(labels map[string]string) string {
	keys, values := stringutils.KeysAndValues(labels)
	name := ""
	for i := range keys {
		name += keys[i] + "-" + values[i] + "-"
	}
	name = strings.TrimSuffix(name, "-")
	return sanitizeName(name)
}

func sanitizeName(name string) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "[", "", -1)
	name = strings.Replace(name, "]", "", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "\n", "", -1)
	return name
}
