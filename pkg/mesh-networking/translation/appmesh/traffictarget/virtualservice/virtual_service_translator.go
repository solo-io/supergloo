package virtualservice

import (
	"context"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
)

type Translator interface {
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		reporter reporting.Reporter,
	) *appmeshv1beta2.VirtualService
}

type translator struct {}
