package istio

import (
	"strings"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
)

func initDestinationRule(writeNamespace, host string, labelSets []map[string]string) *v1alpha3.DestinationRule {

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
	return &v1alpha3.DestinationRule{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      host,
		},
		Host:    host,
		Subsets: subsets,
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
