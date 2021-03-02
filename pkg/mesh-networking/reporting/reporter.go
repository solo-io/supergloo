package reporting

import (
	"context"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

//go:generate mockgen -source ./reporter.go -destination mocks/reporter.go

// the reporter reports status errors on user configuration objects
type Reporter interface {
	// report an error on a TrafficPolicy that has been applied to a Destination
	ReportTrafficPolicyToDestination(destination *discoveryv1.Destination, trafficPolicy ezkube.ResourceId, err error)

	// report an error on an AccessPolicy that has been applied to a Destination
	ReportAccessPolicyToDestination(destination *discoveryv1.Destination, accessPolicy ezkube.ResourceId, err error)

	// report an error on a VirtualMesh that has been applied to a Mesh
	ReportVirtualMeshToMesh(mesh *discoveryv1.Mesh, virtualMesh ezkube.ResourceId, err error)
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

func (p *panickingReporter) ReportTrafficPolicyToDestination(destination *discoveryv1.Destination, trafficPolicy ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw(
			"internal error: error reported on TrafficPolicy which should have been caught by validation!",
			"policy", sets.Key(trafficPolicy),
			"traffic-target", sets.Key(destination),
			"error", err)
}

func (p *panickingReporter) ReportAccessPolicyToDestination(destination *discoveryv1.Destination, accessPolicy ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw("internal error: error reported on AccessPolicy which should have been caught by validation!",
			"policy", sets.Key(accessPolicy),
			"traffic-target", sets.Key(destination),
			"error", err)
}

func (p *panickingReporter) ReportVirtualMeshToMesh(mesh *discoveryv1.Mesh, virtualMesh ezkube.ResourceId, err error) {
	contextutils.LoggerFrom(p.ctx).
		DPanicw("internal error: error reported on VirtualMesh which should have been caught by validation!",
			"mesh", sets.Key(mesh),
			"virtual-mesh", sets.Key(virtualMesh),
			"error", err)
}
