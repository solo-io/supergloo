package virtualservice

import (
	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
)

type Translator interface {
	Translate(
		in input.Snapshot,
	) *v1beta2.VirtualService
}
