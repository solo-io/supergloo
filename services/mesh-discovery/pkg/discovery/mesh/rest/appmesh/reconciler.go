package appmesh

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	appmesh2 "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/aws/parser"
	settings_utils "github.com/solo-io/service-mesh-hub/pkg/aws/selection"
	"github.com/solo-io/service-mesh-hub/pkg/aws/settings"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	compute_target_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconcilerName = "AppMesh reconciler"
)

var (
	NumItemsPerRequest = aws.Int64(100)
)

type appMeshDiscoveryReconciler struct {
	arnParser            aws_utils.ArnParser
	meshClient           zephyr_discovery.MeshClient
	appmeshClientFactory appmesh2.AppmeshRawClientFactory
	settingsClient       settings.SettingsHelperClient
	awsSelector          settings_utils.AwsSelector
}

func NewAppMeshDiscoveryReconciler(
	masterClient client.Client,
	meshClientFactory zephyr_discovery.MeshClientFactory,
	arnParser aws_utils.ArnParser,
	appmeshClientFactory appmesh2.AppmeshRawClientFactory,
	settingsClient settings.SettingsHelperClient,
	awsSelector settings_utils.AwsSelector,
) compute_target_aws.AppMeshDiscoveryReconciler {
	return &appMeshDiscoveryReconciler{
		arnParser:            arnParser,
		meshClient:           meshClientFactory(masterClient),
		appmeshClientFactory: appmeshClientFactory,
		settingsClient:       settingsClient,
		awsSelector:          awsSelector,
	}
}

func (a *appMeshDiscoveryReconciler) GetName() string {
	return ReconcilerName
}

// Currently Meshes are the only SMH CRD that are discovered through the AWS REST API
// For EKS, workloads and services are discovered directly from the cluster.
func (a *appMeshDiscoveryReconciler) Reconcile(ctx context.Context, creds *credentials.Credentials, accountID string) error {
	logger := contextutils.LoggerFrom(ctx)
	selectorsByRegion, err := a.fetchSelectorsByRegion(ctx, accountID)
	if err != nil {
		return err
	}
	// Set containing SMH unique identifiers for AppMesh instances (the Mesh CRD name) for comparison
	discoveredSMHMeshNames := sets.NewString()
	var errors *multierror.Error
	for region, selectors := range selectorsByRegion {
		appmeshClient, err := a.appmeshClientFactory(creds, region)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		var nextToken *string
		input := &appmesh.ListMeshesInput{
			Limit:     NumItemsPerRequest,
			NextToken: nextToken,
		}
		for {
			logger.Debugf("Listing Appmeshes with input %+v", input)
			appMeshes, err := appmeshClient.ListMeshes(input)
			if err != nil {
				errors = multierror.Append(errors, err)
				break
			}
			for _, appMeshRef := range appMeshes.Meshes {
				appMeshTagsOutput, err := appmeshClient.ListTagsForResource(&appmesh.ListTagsForResourceInput{ResourceArn: appMeshRef.Arn})
				if err != nil {
					errors = multierror.Append(errors, err)
					continue
				}
				matched, err := a.awsSelector.AppMeshMatchedBySelectors(appMeshRef, appMeshTagsOutput.Tags, selectors)
				if err != nil {
					return err
				}
				if !matched {
					continue
				}
				discoveredMesh, err := a.convertAppMesh(appMeshRef, region)
				if err != nil {
					return err
				}
				discoveredSMHMeshNames.Insert(discoveredMesh.GetName())
				// Create Mesh only if it doesn't exist to avoid overwriting the clusters field.
				_, err = a.meshClient.GetMesh(
					ctx,
					client.ObjectKey{Name: discoveredMesh.GetName(), Namespace: discoveredMesh.GetNamespace()},
				)
				if k8s_errs.IsNotFound(err) {
					err = a.meshClient.CreateMesh(ctx, discoveredMesh)
					if err != nil {
						return err
					}
				} else if err != nil {
					return err
				}
			}
			if appMeshes.NextToken == nil {
				break
			}
			input.SetNextToken(aws.StringValue(appMeshes.NextToken))
		}
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
	return errors.ErrorOrNil()
}

func (a *appMeshDiscoveryReconciler) convertAppMesh(appMeshRef *appmesh.MeshRef, region string) (*zephyr_discovery.Mesh, error) {
	meshName := metadata.BuildAppMeshName(aws.StringValue(appMeshRef.MeshName), region, aws.StringValue(appMeshRef.MeshOwner))
	awsAccountID, err := a.arnParser.ParseAccountID(aws.StringValue(appMeshRef.Arn))
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.Mesh{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      meshName,
			Namespace: container_runtime.GetWriteNamespace(),
		},
		Spec: zephyr_discovery_types.MeshSpec{
			MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
				AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
					Name:         aws.StringValue(appMeshRef.MeshName),
					AwsAccountId: awsAccountID,
					Region:       region,
				},
			},
		},
	}, nil
}

func (a *appMeshDiscoveryReconciler) fetchSelectorsByRegion(
	ctx context.Context,
	accountID string,
) (settings_utils.AwsSelectorsByRegion, error) {
	awsSettings, err := a.settingsClient.GetAWSSettingsForAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if awsSettings == nil || awsSettings.GetMeshDiscovery().GetDisabled() {
		return nil, nil
	}
	if a.awsSelector.IsDiscoverAll(awsSettings.GetMeshDiscovery()) ||
		(awsSettings.GetMeshDiscovery() != nil && len(awsSettings.GetMeshDiscovery().GetResourceSelectors()) == 0) {
		return a.awsSelector.AwsSelectorsForAllRegions(), nil
	}
	return a.awsSelector.ResourceSelectorsByRegion(awsSettings.GetMeshDiscovery().GetResourceSelectors())
}
