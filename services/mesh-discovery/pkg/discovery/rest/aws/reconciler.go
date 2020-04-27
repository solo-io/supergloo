package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/rotisserie/eris"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ObjectNamePrefix   = "appmesh"
	NumItemsPerRequest = aws.Int64(100)
	ARNParseError      = func(err error, arn string) error {
		return eris.Wrapf(err, "Error parsing ARN: %s", arn)
	}
)

type appMeshDiscoveryReconciler struct {
	meshPlatformName   string
	region             string
	meshClient         zephyr_discovery.MeshClient
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient
	meshServiceClient  zephyr_discovery.MeshServiceClient
	appMeshClient      appmeshiface.AppMeshAPI
}

type AppMeshDiscoveryReconcilerFactory func(
	meshPlatformName string,
	appMeshClient appmeshiface.AppMeshAPI,
	region string,
) rest.RestAPIDiscoveryReconciler

func NewAppMeshDiscoveryReconcilerFactory(
	masterClient client.Client,
	meshClientFactory zephyr_discovery.MeshClientFactory,
) AppMeshDiscoveryReconcilerFactory {
	return func(
		meshPlatformName string,
		appMeshClient appmeshiface.AppMeshAPI,
		region string,
	) rest.RestAPIDiscoveryReconciler {
		return NewAppMeshDiscoveryReconciler(
			meshClientFactory(masterClient),
			appMeshClient,
			meshPlatformName,
			region,
		)
	}
}

func NewAppMeshDiscoveryReconciler(
	meshClient zephyr_discovery.MeshClient,
	appMeshClient appmeshiface.AppMeshAPI,
	meshPlatformName string,
	region string,
) rest.RestAPIDiscoveryReconciler {
	return &appMeshDiscoveryReconciler{
		meshClient:       meshClient,
		appMeshClient:    appMeshClient,
		meshPlatformName: meshPlatformName,
		region:           region,
	}
}

func (a *appMeshDiscoveryReconciler) Reconcile(ctx context.Context) error {
	// Currently Meshes are the only SMH CRD that are discovered through the AWS REST API
	// For EKS, workloads and services are discovered directly from the cluster.
	return a.reconcileMeshes(ctx, a.appMeshClient)
}

func (a *appMeshDiscoveryReconciler) reconcileMeshes(
	ctx context.Context,
	appMeshClient appmeshiface.AppMeshAPI,
) error {
	var nextToken *string
	input := &appmesh.ListMeshesInput{
		Limit:     NumItemsPerRequest,
		NextToken: nextToken,
	}
	discoveredSMHMeshNames := sets.NewString() // Set containing SMH unique identifiers for AppMesh instances (the Mesh CRD name) for comparison
	for {
		appMeshes, err := appMeshClient.ListMeshes(input)
		if err != nil {
			return err
		}
		for _, appMeshRef := range appMeshes.Meshes {
			discoveredMesh, err := a.convertAppMeshMesh(appMeshRef)
			if err != nil {
				return err
			}
			discoveredSMHMeshNames.Insert(discoveredMesh.GetName())
			err = a.meshClient.UpsertMeshSpec(ctx, discoveredMesh)
			if err != nil {
				return err
			}
		}
		if appMeshes.NextToken == nil {
			break
		}
		input.SetNextToken(aws.StringValue(appMeshes.NextToken))
	}

	meshList, err := a.meshClient.ListMesh(ctx)
	if err != nil {
		return err
	}
	for _, mesh := range meshList.Items {
		mesh := mesh
		if mesh.Spec.GetAwsAppMesh() == nil {
			continue
		}
		// Remove Meshes that no longer exist in AppMesh
		if !discoveredSMHMeshNames.Has(mesh.GetName()) {
			err := a.meshClient.DeleteMesh(ctx, client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *appMeshDiscoveryReconciler) convertAppMeshMesh(appMeshRef *appmesh.MeshRef) (*zephyr_discovery.Mesh, error) {
	meshName := a.buildAppMeshMeshName(appMeshRef)
	awsAccountID, err := a.getAwsAccountID(appMeshRef)
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.Mesh{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      meshName,
			Namespace: env.GetWriteNamespace(),
		},
		Spec: zephyr_discovery_types.MeshSpec{
			MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
				AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
					Name:         aws.StringValue(appMeshRef.MeshName),
					AwsAccountId: awsAccountID,
					Region:       a.region,
				},
			},
		},
	}, nil
}

func (a *appMeshDiscoveryReconciler) getAwsAccountID(appMeshRef *appmesh.MeshRef) (string, error) {
	parse, err := arn.Parse(aws.StringValue(appMeshRef.Arn))
	if err != nil {
		return "", ARNParseError(err, aws.StringValue(appMeshRef.Arn))
	}
	return parse.AccountID, nil
}

// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
// Secret name identifies the user-supplied AWS Account registration name
// Format: <mesh_type>-<mesh_entity_name>-<parent_mesh_name>-<aws_account_registration_name>
func (a *appMeshDiscoveryReconciler) buildAppMeshMeshName(mesh *appmesh.MeshRef) string {
	return fmt.Sprintf("%s-%s-%s",
		ObjectNamePrefix,
		convertToKubeName(aws.StringValue(mesh.MeshName)),
		a.meshPlatformName)
}

// AppMesh entity names only contain "Alphanumeric characters, dashes, and underscores are allowed." (taken from AppMesh GUI)
// So just replace underscores with a k8s name friendly delimiter
func convertToKubeName(appmeshName string) string {
	return strings.ReplaceAll(appmeshName, "_", "-")
}
