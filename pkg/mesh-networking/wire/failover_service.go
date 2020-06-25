package wire

import (
	"github.com/google/wire"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/providers"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/providers"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/reconcile"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/validation"
)

var (
	FailoverServiceProviderSet = wire.NewSet(
		reconcile.NewFailoverServiceReconciler,
		reconcile.NewFailoverServiceProcessor,
		validation.NewFailoverServiceValidator,
		v1alpha12.FailoverServiceClientProvider,
		v1alpha1.KubernetesClusterClientProvider,
		FailoverServiceTranslatorProvider,
		istio.NewIstioFailoverServiceTranslator,
	)
)

func FailoverServiceTranslatorProvider(
	istioFailoverServiceTranslator istio.IstioFailoverServiceTranslator,
) []translation.FailoverServiceTranslator {
	return []translation.FailoverServiceTranslator{
		istioFailoverServiceTranslator,
	}
}
