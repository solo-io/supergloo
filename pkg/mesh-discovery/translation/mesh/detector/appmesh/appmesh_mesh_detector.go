package appmesh

import (
	"context"

	"github.com/hashicorp/go-multierror"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	appsv1 "k8s.io/api/apps/v1"
)

type meshDetector struct {
	ctx context.Context
}

func NewMeshDetector(
	ctx context.Context,
) detector.MeshDetector {
	return &meshDetector{
		ctx: ctx,
	}
}

// returns a mesh for each unique AppMesh Controller Mesh CRD in the snapshot
func (d *meshDetector) DetectMeshes(in input.Snapshot) (v1alpha2.MeshSlice, error) {
	var meshes v1alpha2.MeshSlice
	var errs error
	for _, deployment := range in.Deployments().List() {
		mesh, err := d.detectMesh(deployment)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		if mesh == nil {
			continue
		}
		meshes = append(meshes, mesh)
	}
	return meshes, errs
}

func (d *meshDetector) detectMesh(deployment *appsv1.Deployment) (*v1alpha2.Mesh, error) {
	// TODO: implement

	return nil, nil
}
