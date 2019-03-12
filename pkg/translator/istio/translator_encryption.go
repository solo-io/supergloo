package istio

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

const kubeApiserverHost = "kubernetes.default.svc.cluster.local"

func makeDestinationRule(ctx context.Context, writeNamespace, host string, labelSets []map[string]string, enableMtls bool) *v1alpha3.DestinationRule {
	// ensure uniqueness of subset names
	usedSubsetNames := make(map[string]struct{})
	var subsets []*v1alpha3.Subset
	for _, labels := range labelSets {
		if len(labels) == 0 {
			continue
		}
		name := utils.SubsetName(labels)
		if _, used := usedSubsetNames[name]; used {
			contextutils.LoggerFrom(ctx).Errorf("internal error: generated name for destination rule "+
				"%v had duplicate subset name %v", host, name)
			continue
		}
		usedSubsetNames[name] = struct{}{}
		subsets = append(subsets, &v1alpha3.Subset{
			Name:   name,
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

		// special case: enable traffic leaving the mesh to the local kube apiserver
		// https://istio.io/docs/tasks/security/authn-policy/#request-from-istio-services-to-kubernetes-api-server
		if host == kubeApiserverHost {
			trafficPolicy.Tls.Mode = v1alpha3.TLSSettings_DISABLE
		}
	}
	return &v1alpha3.DestinationRule{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      utils.SanitizeName(host),
		},
		Host:          host,
		Subsets:       subsets,
		TrafficPolicy: trafficPolicy,
	}
}
