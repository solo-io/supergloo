package appmesh

import (
	"context"
	"fmt"
	"sort"
	"strings"

	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"

	aws_v1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/labelutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// The prefix for the mesh resource type in an ARN.
	meshResourcePrefix = "mesh/"
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
func (d *meshDetector) DetectMeshes(in input.DiscoveryInputSnapshot, _ *settingsv1.DiscoverySettings) (v1.MeshSlice, error) {
	var errors error

	// Group meshes by ARN because meshes that share an ARN are backed by the same AWS resources.
	awsMeshByArn := make(map[string][]*aws_v1beta2.Mesh)
	for _, awsMesh := range in.Meshes().List() {
		if awsMesh.Status.MeshARN == nil {
			// Meshes that lack an ARN have not been processed by the App Mesh controller; ignore.
			continue
		}

		awsMeshByArn[*awsMesh.Status.MeshARN] = append(awsMeshByArn[*awsMesh.Status.MeshARN], awsMesh)
	}

	// Produce a sorted ARN list to ensure map iteration is consistent across reconciles.
	var arnList []string
	for meshArn := range awsMeshByArn {
		arnList = append(arnList, meshArn)
	}
	sort.Strings(arnList)

	// Create one discovery artifact for each ARN's mesh list.
	var output v1.MeshSlice
	for _, meshArn := range arnList {
		discoveredMesh, err := d.discoverMesh(meshArn, awsMeshByArn[meshArn])
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		output = append(output, discoveredMesh)
	}

	return output, errors
}

func (d *meshDetector) discoverMesh(meshArn string, awsMeshList []*aws_v1beta2.Mesh) (*v1.Mesh, error) {
	parsedArn, err := arn.Parse(meshArn)
	if err != nil {
		return nil, err
	}

	meshName := strings.TrimPrefix(parsedArn.Resource, meshResourcePrefix)

	clusters := make([]string, 0, len(awsMeshList))
	for _, awsMesh := range awsMeshList {
		clusters = append(clusters, awsMesh.ClusterName)
	}
	sort.Strings(clusters)

	mesh := &v1.Mesh{
		ObjectMeta: discoveredMeshObjectMeta(meshName, parsedArn.Region, parsedArn.AccountID),
		Spec: v1.MeshSpec{
			Type: &v1.MeshSpec_AwsAppMesh_{
				AwsAppMesh: &v1.MeshSpec_AwsAppMesh{
					AwsName:      meshName,
					Region:       parsedArn.Region,
					AwsAccountId: parsedArn.AccountID,
					Arn:          meshArn,
					Clusters:     clusters,
				},
			},
		},
	}
	return mesh, nil
}

// discoveredMeshObjectMeta returns ObjectMeta for a discovered AWS App Mesh instance.
// This differs from utils.DiscoveredObjectMeta because App Mesh mesh resources have no namespace
// and can appear on any number of Kubernetes clusters.
func discoveredMeshObjectMeta(meshName, region, accountID string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: defaults.GetPodNamespace(),
		Name:      fmt.Sprintf("%s-%s-%s", meshName, region, accountID),
		Labels:    labelutils.OwnershipLabels(),
	}
}
