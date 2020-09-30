package appmesh

import (
	"context"
	"strings"

	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
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
	var meshes v1alpha2.MeshSlice
	var errs error

	for _, awsMesh := range in.Meshes().List() {
		mesh, err := d.detectMesh(awsMesh)
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

func (d *meshDetector) detectMesh(mesh *v1beta2.Mesh) (*v1alpha2.Mesh, error) {
	if mesh.Status.MeshARN == nil || mesh.Spec.AWSName == nil {
		return nil, nil
	}

	region, accountId, _, err := parseArn(*mesh.Status.MeshARN)
	if err != nil {
		return nil, err
	}

	return &v1alpha2.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(mesh),
		Spec: v1alpha2.MeshSpec{
			MeshType: &v1alpha2.MeshSpec_AwsAppMesh_{
				AwsAppMesh: &v1alpha2.MeshSpec_AwsAppMesh{
					AwsName:      *mesh.Spec.AWSName,
					Region:       region,
					AwsAccountId: accountId,
					Arn:          *mesh.Status.MeshARN,
					// TODO -- look at pods, APPMESH_VIRTUALNODE_NAME will contain mesh name
					Clusters: []string{mesh.ClusterName},
				},
			},
		},
	}, nil
}

// parseArn extracts mesh name, region, and account ID from a mesh ARN.
func parseArn(arn string) (region string, accountId string, meshName string, err error) {
	// Value takes format "arn:aws:appmesh:<region>:<account ID>:mesh/<mesh name>"
	colonSplit := strings.Split(arn, ":")
	if len(colonSplit) != 6 {
		return "", "", "", eris.Errorf("Unexpected mesh ARN %s", arn)
	}
	slashSplit := strings.Split(arn, "/")
	if len(slashSplit) != 2 {
		return "", "", "", eris.Errorf("Unexpected mesh ARN %s", arn)
	}

	region = colonSplit[3]
	accountId = colonSplit[4]
	meshName = slashSplit[1]
	return
}
