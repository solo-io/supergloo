package reporter

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// the reporter reports status errors on user configuration objects
type Reporter interface {
	// report an error on a traffic policy that has been applied to a MeshService
	ReportTrafficPolicy(meshService *v1alpha1.MeshService, trafficPolicy ezkube.ResourceId, err error)

	// report an error on an access policy that has been applied to a MeshService
	ReportAccessPolicy(meshService *v1alpha1.MeshService, accessPolicy ezkube.ResourceId, err error)
}
