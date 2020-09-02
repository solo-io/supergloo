package appmesh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/external/appmesh"
	"github.com/solo-io/skv2/contrib/pkg/output"
)

//go:generate mockgen -source ./appmesh_output_syncer.go -destination mocks/mock_appmesh_output_syncer.go

// the appmesh OutputSyncer handles syncing appmesh snapshots to
// the Appmesh API hosted on AWS.
type OutputSyncer interface {
	Apply(ctx context.Context, outputs appmesh.Snapshot, errHandler output.ErrorHandler) error
}

type outputSyncer struct {
	appmeshClient appmesh.Client
}

func NewOutputSyncer(appmeshClient appmesh.Client) OutputSyncer {
	return &outputSyncer{appmeshClient: appmeshClient}
}

func (s *outputSyncer) Apply(ctx context.Context, outputs appmesh.Snapshot, errHandler output.ErrorHandler) error {
	for _, mesh := range outputs.Meshes() {
		if err := s.syncMesh(ctx, outputs, mesh, errHandler); err != nil {
			return err
		}
	}

	return nil
}

/*
The order for syncing a mesh as follows:
* Upsert desired VirtualNodes
* Upsert desired VirtualRouters
* Upsert desired Routes
* Upsert desired VirtualServices
* Delete stale VirtualServices
* Delete stale Routes
* Delete stale VirtualRouters
* Delete stale VirtualNodes
*/
func (s *outputSyncer) syncMesh(ctx context.Context, outputs appmesh.Snapshot, mesh string, errHandler output.ErrorHandler) error {
	existingVirtualNodes, err := s.appmeshClient.ListVirtualNodes(ctx, mesh)
	if err != nil {
		return err
	}

	existingVirtualRouters, err := s.appmeshClient.ListVirtualRouters(ctx, mesh)
	if err != nil {
		return err
	}

	existingRoutes, err := s.appmeshClient.ListRoutes(ctx, mesh)
	if err != nil {
		return err
	}

	existingVirtualServices, err := s.appmeshClient.ListVirtualServices(ctx, mesh)
	if err != nil {
		return err
	}

	s.appmeshClient.UpsertVirtualNodes(ctx, mesh, existingVirtualNodes, outputs.VirtualNodes().List(), errHandler)
	s.appmeshClient.UpsertVirtualRouters(ctx, mesh, existingVirtualRouters, outputs.VirtualRouters().List(), errHandler)
	s.appmeshClient.UpsertRoutes(ctx, mesh, existingRoutes, outputs.Routes().List(), errHandler)
	s.appmeshClient.UpsertVirtualServices(ctx, mesh, existingVirtualServices, outputs.VirtualServices().List(), errHandler)

	s.appmeshClient.DeleteVirtualServices(ctx, mesh, existingVirtualServices, errHandler)
	s.appmeshClient.DeleteRoutes(ctx, mesh, existingRoutes, errHandler)
	s.appmeshClient.DeleteVirtualRouters(ctx, mesh, existingVirtualRouters, errHandler)
	s.appmeshClient.DeleteVirtualNodes(ctx, mesh, existingVirtualNodes, errHandler)

	return nil
}
