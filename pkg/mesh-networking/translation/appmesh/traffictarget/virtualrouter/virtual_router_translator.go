package virtualrouter

import (
	"context"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
)

type Translator interface {
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		reporter reporting.Reporter,
	) *appmeshv1beta2.VirtualRouter
}

type translator struct{}

func NewVirtualRouterTranslator() Translator {
	return &translator{}
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) *appmeshv1beta2.VirtualRouter {

	kubeService := trafficTarget.Spec.GetKubeService()
	if kubeService == nil {
		// TODO non kube services currently unsupported
		return nil
	}

	routeTranslator := newRouteTranslator(reporter, in.TrafficTargets(), in.Workloads())
	routes := routeTranslator.getRoutes(trafficTarget)
	if len(routes) == 0 {
		// There are no routes, so we don't need to create a virtual router
		return nil
	}

	listenerTranslator := newListenerTranslator()
	listeners := listenerTranslator.getListeners(trafficTarget)

	return &appmeshv1beta2.VirtualRouter{
		ObjectMeta: metautils.TranslatedObjectMeta(
			trafficTarget.Spec.GetKubeService().Ref,
			trafficTarget.Annotations,
		),
		Spec: appmeshv1beta2.VirtualRouterSpec{
			Listeners: listeners,
			Routes:    routes,
		},
	}
}
