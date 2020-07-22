package reporting

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// the reporter reports status errors on user configuration objects
type Reporter interface {
	// report an error on a traffic policy that has been applied to a MeshService
	ReportTrafficPolicy(meshService *v1alpha2.MeshService, trafficPolicy ezkube.ResourceId, err error)

	// report an error on an access policy that has been applied to a MeshService
	ReportAccessPolicy(meshService *v1alpha2.MeshService, accessPolicy ezkube.ResourceId, err error)

	// report an error on a virtual mesh that has been applied to a Mesh
	ReportVirtualMesh(mesh *v1alpha2.Mesh, virtualMesh ezkube.ResourceId, err error)

	// report an error on a failover service that has been applied to a Mesh
	ReportFailoverService(mesh *v1alpha2.Mesh, failoverService ezkube.ResourceId, err error)
}

// this reporter implementation is only used inside
// the real translation, which translates a validated snapshot.
// therefore, reports should only ever occur if we have a bug, in which case this reporter will issue a DPanic log (panicking in development mode)
type panickingReporter struct {
	ctx context.Context
}

func NewPanickingReporter(ctx context.Context) Reporter {
	return &panickingReporter{ctx: ctx}
}

func (p *panickingReporter) ReportTrafficPolicy(meshService *v1alpha2.MeshService, trafficPolicy ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw(
			"internal error: error reported on TrafficPolicy which should have been caught by validation!",
			"policy", sets.Key(trafficPolicy),
			"mesh-service", sets.Key(meshService),
			"error", err)
}

func (p *panickingReporter) ReportAccessPolicy(meshService *v1alpha2.MeshService, accessPolicy ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw("internal error: error reported on AccessPolicy which should have been caught by validation!",
			"policy", sets.Key(accessPolicy),
			"mesh-service", sets.Key(meshService),
			"error", err)
}

func (p *panickingReporter) ReportVirtualMesh(mesh *v1alpha2.Mesh, virtualMesh ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw("internal error: error reported on VirtualMesh which should have been caught by validation!",
			"mesh", sets.Key(mesh),
			"virtual-mesh", sets.Key(virtualMesh),
			"error", err)
}

func (p *panickingReporter) ReportFailoverService(mesh *v1alpha2.Mesh, failoverService ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw("internal error: error reported on FailoverService which should have been caught by validation!",
			"mesh", sets.Key(mesh),
			"failover-service", sets.Key(failoverService),
			"error", err)
}
