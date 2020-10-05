package appmesh

import (
	"context"
	"sort"

	aws_v1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
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
	return d.detectMeshes(in.Meshes().List())
}

func (d *meshDetector) detectMeshes(meshList []*aws_v1beta2.Mesh) (v1alpha2.MeshSlice, error) {
	var errs error

	// Meshes that have the same ARN will be treated as identical.
	discoveredMeshByARN := make(map[string]*v1alpha2.Mesh)
	for _, awsMesh := range meshList {
		if awsMesh.Status.MeshARN == nil {
			// Meshes that lack an ARN have not been processed by the App Mesh controller; ignore for now.
			continue
		}
		if mesh, found := discoveredMeshByARN[*awsMesh.Status.MeshARN]; found {
			// We have seen a mesh with this ARN. Record the awsMesh's cluster to our corresponding mesh record.
			mesh.Spec.GetAwsAppMesh().Clusters = append(mesh.Spec.GetAwsAppMesh().Clusters, awsMesh.ClusterName)
		} else {
			// We have not seen a mesh with this ARN, create a new mesh record.
			discoveredMesh, err := d.discoverNewMesh(awsMesh)
			if err != nil {
				errs = multierror.Append(errs, err)
			} else {
				discoveredMeshByARN[*awsMesh.Status.MeshARN] = discoveredMesh
			}
		}
	}

	// Sort the cluster lists on each mesh resource for idempotence.
	output := make(v1alpha2.MeshSlice, 0, len(discoveredMeshByARN))
	for _, discoveredMesh := range discoveredMeshByARN {
		discoveredMesh := discoveredMesh
		sort.Strings(discoveredMesh.Spec.GetAwsAppMesh().Clusters)
		output = append(output, discoveredMesh)
	}
	return output, errs
}

func (d *meshDetector) discoverNewMesh(awsMesh *aws_v1beta2.Mesh) (*v1alpha2.Mesh, error) {
	parsedArn, err := arn.Parse(*awsMesh.Status.MeshARN)
	if err != nil {
		return nil, err
	}

	var meshName string
	if awsMesh.Spec.AWSName != nil {
		meshName = *awsMesh.Spec.AWSName
	} else {
		meshName = awsMesh.Name
	}

	return &v1alpha2.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(awsMesh),
		Spec: v1alpha2.MeshSpec{
			MeshType: &v1alpha2.MeshSpec_AwsAppMesh_{
				AwsAppMesh: &v1alpha2.MeshSpec_AwsAppMesh{
					AwsName:      meshName,
					Region:       parsedArn.Region,
					AwsAccountId: parsedArn.AccountID,
					Arn:          *awsMesh.Status.MeshARN,
					Clusters:     []string{awsMesh.ClusterName},
				},
			},
		},
	}, nil
}
