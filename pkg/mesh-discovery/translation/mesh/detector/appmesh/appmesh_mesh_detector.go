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
	// Meshes that have the same ARN refer to the same entity in AWS.
	discoveredMeshByARN := make(map[string]*v1alpha2.Mesh)
	var errs error
	for _, awsMesh := range meshList {
		if awsMesh.Status.MeshARN == nil {
			// Meshes that lack an ARN have not been processed by the App Mesh controller; ignore.
			continue
		}

		if mesh, found := discoveredMeshByARN[*awsMesh.Status.MeshARN]; found {
			// We have seen a mesh with this ARN.
			// Add this awsMesh's cluster to the list of clusters the mesh configures.
			mesh.Spec.GetAwsAppMesh().Clusters = append(mesh.Spec.GetAwsAppMesh().Clusters, awsMesh.ClusterName)
		} else {
			// We have not seen a mesh with this ARN.
			// Create a new mesh record.
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
	// If AWSName is not set, fallback to metadata.Name per https://github.com/aws/aws-app-mesh-controller-for-k8s/blob/v1.1.1/apis/appmesh/v1beta2/mesh_types.go#L66
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
