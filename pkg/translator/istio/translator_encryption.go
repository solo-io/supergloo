package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
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
	encryptionRules v1.EncryptionRuleList,
	resourceErrs reporter.ResourceErrors,
) *v1alpha3.DestinationRule {
	dr := initDestinationRule(t.writeNamespace, host, labelSets)

	for _, er := range encryptionRules {

		// add a rule for each dest port
		for _, plug := range t.plugins {
			encryptionPlugin, ok := plug.(plugins.EncryptionPlugin)
			if !ok {
				continue
			}
			if er.Spec == nil {
				resourceErrs.AddError(er, errors.Errorf("spec cannot be empty"))
				continue
			}
			if err := encryptionPlugin.ProcessDestinationRule(params, *er.Spec, dr); err != nil {
				resourceErrs.AddError(er, errors.Wrapf(err, "applying destination rule failed"))
			}
		}
	}

	return dr
}

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
