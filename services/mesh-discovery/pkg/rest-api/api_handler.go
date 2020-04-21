package rest_api

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/rest_watcher"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	NumItemsPerRequest = aws.Int64(100)
)

type appMeshAPIReconciler struct {
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter
	secret                  *k8s_core_types.Secret
	meshClient              zephyr_discovery.MeshClient
	meshWorkloadClient      zephyr_discovery.MeshWorkloadClient
	meshServiceClient       zephyr_discovery.MeshServiceClient
}

type AppMeshReconcilerFactory func(secret *k8s_core_types.Secret) rest_watcher.RestAPIReconciler

func NewAppMeshAPIReconcilerFactory(
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter,
	masterClient client.Client,
	meshClientFactory zephyr_discovery.MeshClientFactory,
	meshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory,
	meshServiceClientFactory zephyr_discovery.MeshServiceClientFactory,
) AppMeshReconcilerFactory {
	return func(
		secret *k8s_core_types.Secret,
	) rest_watcher.RestAPIReconciler {
		return NewAppMeshAPIReconciler(
			secretAwsCredsConverter,
			meshClientFactory(masterClient),
			meshWorkloadClientFactory(masterClient),
			meshServiceClientFactory(masterClient),
			secret,
		)
	}
}

func NewAppMeshAPIReconciler(
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter,
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	secret *k8s_core_types.Secret,
) rest_watcher.RestAPIReconciler {
	return &appMeshAPIReconciler{
		secretAwsCredsConverter: secretAwsCredsConverter,
		meshClient:              meshClient,
		meshWorkloadClient:      meshWorkloadClient,
		meshServiceClient:       meshServiceClient,
		secret:                  secret,
	}
}

func (a *appMeshAPIReconciler) Reconcile(ctx context.Context) error {
	var err error
	var appMeshClient *appmesh.AppMesh
	if appMeshClient, err = a.buildAppMeshClient(a.secret); err != nil {
		return err
	}
	if err = a.reconcileMeshes(ctx, appMeshClient); err != nil {
		return err
	}
	if err = a.reconcileMeshWorkloads(appMeshClient); err != nil {
		return err
	}
	if err = a.reconcileMeshServices(appMeshClient); err != nil {
		return err
	}
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshes(ctx context.Context, appMeshClient *appmesh.AppMesh) error {
	var nextToken *string
	input := &appmesh.ListMeshesInput{
		Limit:     NumItemsPerRequest,
		NextToken: nextToken,
	}
	var meshes []*zephyr_discovery.Mesh
	// Set containing SMH unique identifiers for AppMesh instances (the Mesh CRD name)
	existingAppMeshNames := sets.NewString()
	for {
		appMeshes, err := appMeshClient.ListMeshes(input)
		if err != nil {
			return err
		}
		if appMeshes.NextToken == nil {
			break
		}
		for _, appMesh := range appMeshes.Meshes {
			meshName := a.buildAppMeshName(appMesh)
			existingAppMeshNames.Insert(meshName)
			meshes = append(meshes, &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      meshName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
						AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
							Name:    aws.StringValue(appMesh.MeshName),
							ApiName: a.secret.GetName(),
							// TODO (harvey) remove hardcode
							Region: "us-east-2",
						},
					},
				},
			})
		}
		input.SetNextToken(*appMeshes.NextToken)
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
		if !existingAppMeshNames.Has(mesh.GetName()) {
			err := a.meshClient.DeleteMesh(ctx, client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()})
			if err != nil {
				return err
			}
		}
		// None of the AwsAppMesh metadata is mutable, so we don't need to update
	}
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshWorkloads(appMeshClient *appmesh.AppMesh) error {
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshServices(appMeshClient *appmesh.AppMesh) error {
	return nil
}

func (a *appMeshAPIReconciler) buildAppMeshClient(secret *k8s_core_types.Secret) (*appmesh.AppMesh, error) {
	creds, err := a.secretAwsCredsConverter.SecretToCreds(secret)
	if err != nil {
		return nil, err
	}
	// TODO (harvey) make region configurable? or scan all regions?
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("us-east-2"),
	})
	if err != nil {
		return nil, err
	}
	return appmesh.New(sess), nil
}

func (a *appMeshAPIReconciler) buildAppMeshName(appmesh *appmesh.MeshRef) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return "appmesh-" + *appmesh.MeshName + a.secret.GetName()
}
